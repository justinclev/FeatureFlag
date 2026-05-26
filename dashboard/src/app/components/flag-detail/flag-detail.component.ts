import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, FormArray, Validators } from '@angular/forms';
import { FlagService } from '../../services/flag.service';
import { Flag, Rule, EvaluationResult } from '../../models/flag.model';

@Component({
  selector: 'app-flag-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, ReactiveFormsModule, RouterLink],
  template: `
    <div class="detail-layout">
      <div class="editor-section">
        <div class="header">
          <a routerLink="/flags" class="back-link">← Back to flags</a>
          <h1>{{ isNew ? 'Create' : 'Edit' }} Feature Flag</h1>
        </div>

        <form [formGroup]="flagForm" (ngSubmit)="save()" class="card">
          <div class="form-group">
            <label>Name</label>
            <input formControlName="name" placeholder="e.g. New Beta Feature">
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Key</label>
              <input formControlName="key" placeholder="e.g. new-beta-feature">
            </div>
            <div class="form-group">
              <label>Strategy</label>
              <select formControlName="ruleMatchStrategy">
                <option value="any">ANY (OR)</option>
                <option value="all">ALL (AND)</option>
              </select>
            </div>
          </div>

          <div class="form-group checkbox-group">
            <label class="switch">
              <input type="checkbox" formControlName="enabled">
              <span class="slider round"></span>
            </label>
            <span>Enabled</span>
          </div>

          <div class="form-group checkbox-group">
            <label class="switch">
              <input type="checkbox" formControlName="defaultValue">
              <span class="slider round"></span>
            </label>
            <span>Default Value (when no rules match)</span>
          </div>

          <div class="rules-header">
            <h3>Rules</h3>
            <button type="button" class="btn btn-sm" (click)="addRule()">+ Add Rule</button>
          </div>

          <div formArrayName="rules" class="rules-list">
            @for (rule of rules.controls; track $index) {
              <div [formGroupName]="$index" class="rule-item">
                <div class="rule-top">
                  <select formControlName="type">
                    <option value="user_list">User List</option>
                    <option value="attribute">Attribute</option>
                    <option value="percentage">Percentage</option>
                  </select>
                  <label class="switch-small">
                    <input type="checkbox" formControlName="value">
                    <span class="slider-small round"></span>
                  </label>
                  <span class="rule-val-label">{{ rule.get('value')?.value ? 'Permit' : 'Deny' }}</span>
                  <button type="button" class="btn-icon delete-btn" (click)="removeRule($index)">×</button>
                </div>
                
                <div class="rule-config" [formGroupName]="'config'">
                  @if (rule.get('type')?.value === 'user_list') {
                    <input formControlName="userIds" placeholder="User IDs (comma separated)">
                  }
                  @if (rule.get('type')?.value === 'attribute') {
                    <div class="attr-row">
                      <input formControlName="attributeKey" placeholder="Key">
                      <select formControlName="attributeOp">
                        <option value="eq">Equals</option>
                        <option value="neq">Not Equals</option>
                        <option value="contains">Contains</option>
                      </select>
                      <input formControlName="attributeValue" placeholder="Value">
                    </div>
                  }
                  @if (rule.get('type')?.value === 'percentage') {
                    <div class="range-group">
                      <input type="range" formControlName="percentage" min="0" max="100">
                      <span>{{ rule.get('config.percentage')?.value }}%</span>
                    </div>
                  }
                </div>
              </div>
            }
          </div>

          <div class="form-actions">
            <button type="submit" class="btn btn-primary" [disabled]="saving()">
              {{ saving() ? 'Saving...' : 'Save Changes' }}
            </button>
            @if (!isNew) {
              <button type="button" class="btn btn-danger" (click)="delete()">Delete</button>
            }
          </div>
        </form>
      </div>

      <div class="test-section">
        <h2>Try it out</h2>
        <div class="card">
          <p class="text-muted">Simulate evaluation with custom context.</p>
          <div class="form-group">
            <label>User ID</label>
            <input [(ngModel)]="testContext.userId" placeholder="e.g. user-123">
          </div>
          <div class="form-group">
            <label>Attributes (JSON)</label>
            <textarea [(ngModel)]="testAttributes" rows="5" placeholder='{"plan": "premium"}'></textarea>
          </div>
          <button class="btn btn-secondary" (click)="test()">Run Evaluation</button>

          @if (testResult()) {
            <div class="test-result" [ngClass]="testResult()?.enabled ? 'res-success' : 'res-fail'">
              <div class="res-head">
                <strong>{{ testResult()?.enabled ? 'ENABLED' : 'DISABLED' }}</strong>
              </div>
              <div class="res-body">{{ testResult()?.reason }}</div>
            </div>
          }
        </div>
      </div>
    </div>
  `,
  styles: [`
    .detail-layout {
      display: grid;
      grid-template-columns: 1fr 320px;
      gap: var(--space-xl);
      align-items: start;
    }
    .header { margin-bottom: var(--space-lg); }
    .back-link {
      display: block;
      color: var(--text-muted);
      text-decoration: none;
      margin-bottom: var(--space-sm);
      font-size: 0.875rem;
    }
    .form-group { margin-bottom: var(--space-md); }
    .form-group label {
      display: block;
      font-size: 0.875rem;
      font-weight: 500;
      color: var(--text-muted);
      margin-bottom: 4px;
    }
    .form-row {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: var(--space-md);
    }
    .checkbox-group {
      display: flex;
      align-items: center;
      gap: var(--space-sm);
      margin-top: var(--space-md);
    }
    .rules-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin: var(--space-xl) 0 var(--space-md);
    }
    .rule-item {
      background: #f8fafc;
      border: 1px solid var(--border);
      border-radius: 6px;
      padding: var(--space-md);
      margin-bottom: var(--space-md);
    }
    .rule-top {
      display: flex;
      align-items: center;
      gap: var(--space-md);
      margin-bottom: var(--space-md);
    }
    .rule-val-label { font-size: 0.875rem; font-weight: 500; min-width: 50px; }
    .btn {
      padding: 0.5rem 1rem;
      border-radius: 6px;
      font-weight: 500;
    }
    .btn-sm { padding: 0.25rem 0.75rem; font-size: 0.875rem; background: var(--bg-main); border: 1px solid var(--border); }
    .btn-primary { background: var(--primary); color: white; width: 100%; }
    .btn-secondary { background: var(--text-main); color: white; width: 100%; margin-top: var(--space-md); }
    .btn-danger { background: white; color: var(--danger); border: 1px solid var(--danger); }
    .btn-danger:hover { background: #fef2f2; }
    .btn-icon { background: none; font-size: 1.25rem; color: var(--text-muted); }
    .delete-btn { margin-left: auto; }
    .delete-btn:hover { color: var(--danger); }
    
    .attr-row { display: grid; grid-template-columns: 1fr 120px 1fr; gap: 8px; }
    .range-group { display: flex; align-items: center; gap: 12px; }

    /* Switch styling */
    .switch { position: relative; display: inline-block; width: 44px; height: 24px; }
    .switch input { opacity: 0; width: 0; height: 0; }
    .slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #ccc; transition: .4s; }
    .slider:before { position: absolute; content: ""; height: 18px; width: 18px; left: 3px; bottom: 3px; background-color: white; transition: .4s; }
    input:checked + .slider { background-color: var(--success); }
    input:checked + .slider:before { transform: translateX(20px); }
    .slider.round { border-radius: 24px; }
    .slider.round:before { border-radius: 50%; }

    .switch-small { position: relative; display: inline-block; width: 34px; height: 18px; }
    .switch-small input { opacity: 0; width: 0; height: 0; }
    .slider-small { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #ef4444; transition: .4s; }
    .slider-small:before { position: absolute; content: ""; height: 12px; width: 12px; left: 3px; bottom: 3px; background-color: white; transition: .4s; }
    input:checked + .slider-small { background-color: var(--success); }
    input:checked + .slider-small:before { transform: translateX(16px); }
    .slider-small.round { border-radius: 18px; }
    .slider-small.round:before { border-radius: 50%; }

    .test-result {
      margin-top: var(--space-xl);
      padding: var(--space-md);
      border-radius: 6px;
      border-left: 4px solid #ccc;
    }
    .res-success { background: #f0fdf4; border-color: var(--success); }
    .res-success strong { color: #166534; }
    .res-fail { background: #fef2f2; border-color: var(--danger); }
    .res-fail strong { color: #991b1b; }
    .res-head { font-size: 0.75rem; margin-bottom: 4px; }
    .res-body { font-size: 0.875rem; color: var(--text-main); }
  `]
})
export class FlagDetailComponent implements OnInit {
  isNew = true;
  flagId: string | null = null;
  flagForm: FormGroup;
  saving = signal(false);
  
  testContext: any = { userId: '', attributes: {} };
  testAttributes = '{}';
  testResult = signal<EvaluationResult | null>(null);

  constructor(
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private flagService: FlagService
  ) {
    this.flagForm = this.fb.group({
      name: ['', Validators.required],
      key: ['', Validators.required],
      description: [''],
      enabled: [true],
      defaultValue: [false],
      ruleMatchStrategy: ['any'],
      rules: this.fb.array([])
    });
  }

  ngOnInit() {
    this.flagId = this.route.snapshot.paramMap.get('id');
    if (this.flagId && this.flagId !== 'new') {
      this.isNew = false;
      this.flagService.getFlag(this.flagId).subscribe(flag => {
        this.patchFlag(flag);
      });
    }
  }

  get rules() { return this.flagForm.get('rules') as FormArray; }

  addRule() {
    const ruleForm = this.fb.group({
      type: ['user_list'],
      value: [true],
      config: this.fb.group({
        userIds: [''],
        attributeKey: [''],
        attributeOp: ['eq'],
        attributeValue: [''],
        percentage: [50]
      })
    });
    this.rules.push(ruleForm);
  }

  removeRule(index: number) { this.rules.removeAt(index); }

  patchFlag(flag: Flag) {
    this.flagForm.patchValue({
      name: flag.name,
      key: flag.key,
      description: flag.description,
      enabled: flag.enabled,
      defaultValue: flag.defaultValue,
      ruleMatchStrategy: flag.ruleMatchStrategy
    });
    
    flag.rules?.forEach(r => {
      const config = { ...r.config };
      if (r.type === 'user_list' && Array.isArray(config.userIds)) {
        config.userIds = config.userIds.join(', ');
      }
      
      const ruleForm = this.fb.group({
        type: [r.type],
        value: [r.value],
        config: this.fb.group({
          userIds: [config.userIds || ''],
          attributeKey: [config.attributeKey || ''],
          attributeOp: [config.attributeOp || 'eq'],
          attributeValue: [config.attributeValue || ''],
          percentage: [config.percentage || 50]
        })
      });
      this.rules.push(ruleForm);
    });
  }

  save() {
    if (this.flagForm.invalid) return;
    this.saving.set(true);
    
    const formVal = this.flagForm.value;
    const rules = formVal.rules.map((r: any) => {
      const config: any = {};
      if (r.type === 'user_list') {
        config.userIds = r.config.userIds.split(',').map((s: string) => s.trim()).filter((s: string) => s);
      } else if (r.type === 'attribute') {
        config.attributeKey = r.config.attributeKey;
        config.attributeOp = r.config.attributeOp;
        config.attributeValue = r.config.attributeValue;
      } else if (r.type === 'percentage') {
        config.percentage = parseFloat(r.config.percentage);
      }
      return { type: r.type, value: r.value, config };
    });

    const flagData = { ...formVal, rules };

    const obs = this.isNew 
      ? this.flagService.createFlag(flagData) 
      : this.flagService.updateFlag(this.flagId!, flagData);

    obs.subscribe({
      next: () => this.router.navigate(['/flags']),
      error: (err) => {
        alert(err.error?.error || 'Failed to save flag');
        this.saving.set(false);
      }
    });
  }

  delete() {
    if (confirm('Are you sure you want to delete this flag?')) {
      this.flagService.deleteFlag(this.flagId!).subscribe(() => {
        this.router.navigate(['/flags']);
      });
    }
  }

  test() {
    try {
      this.testContext.attributes = JSON.parse(this.testAttributes);
    } catch (e) {
      alert('Invalid JSON in attributes');
      return;
    }

    this.flagService.evaluate(this.flagForm.get('key')?.value, this.testContext).subscribe({
      next: (res) => this.testResult.set(res),
      error: (err) => alert(err.error?.error || 'Evaluation failed')
    });
  }
}
