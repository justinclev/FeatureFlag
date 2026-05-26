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

        <form [formGroup]="flagForm" (ngSubmit)="save()">
          <!-- Section 1: Identity -->
          <div class="card mb-lg">
            <div class="section-title">
              <span class="step-badge">1</span>
              <h3>Identity & Strategy</h3>
            </div>
            <div class="form-group">
              <label>Flag Name</label>
              <input formControlName="name" placeholder="e.g. New Beta Feature">
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>Key</label>
                <input formControlName="key" placeholder="e.g. new-beta-feature">
              </div>
              <div class="form-group">
                <label>Rule Match Strategy</label>
                <select formControlName="ruleMatchStrategy">
                  <option value="any">ANY (Short-circuit / Deny Wins)</option>
                  <option value="all">ALL (Must match every rule)</option>
                </select>
              </div>
            </div>
          </div>

          <!-- Section 2: Safety & Master Switch -->
          <div class="card mb-lg highlight-card" [ngClass]="{'card-off': !flagForm.get('enabled')?.value}">
            <div class="section-title">
              <span class="step-badge">2</span>
              <h3>Master Control & Safety</h3>
            </div>
            <div class="control-grid">
              <div class="control-item">
                <label class="switch-large">
                  <input type="checkbox" formControlName="enabled">
                  <span class="slider-large round"></span>
                </label>
                <div class="label-group">
                  <span class="control-label">{{ flagForm.get('enabled')?.value ? 'ENABLED' : 'DISABLED' }}</span>
                  <span class="text-xs text-muted">Master toggle for this feature flag.</span>
                </div>
              </div>
              <div class="control-item">
                <label class="switch">
                  <input type="checkbox" formControlName="offValue">
                  <span class="slider round"></span>
                </label>
                <div class="label-group">
                  <span class="font-medium">Off Value</span>
                  <span class="text-xs text-muted">Constant value returned when master toggle is OFF.</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Section 3: Targeting (Only if enabled) -->
          <div class="card" [ngClass]="{'dimmed': !flagForm.get('enabled')?.value}">
            <div class="section-title">
              <span class="step-badge">3</span>
              <h3>Targeting Rules</h3>
              @if (!flagForm.get('enabled')?.value) {
                <span class="pill pill-muted ml-auto">Inactive</span>
              }
            </div>
            
            <div class="rules-container">
              <div formArrayName="rules" class="rules-list">
                @for (rule of rules.controls; track $index) {
                  <div [formGroupName]="$index" class="rule-item">
                    <!-- ... existing rule item content ... -->
                    <div class="rule-top">
                      <select formControlName="type">
                        <option value="user_list">User List</option>
                        <option value="attribute">Attribute</option>
                        <option value="percentage">Percentage</option>
                        <option value="geography">Geography</option>
                        <option value="gradual">Gradual Rollout</option>
                        <option value="schedule">Time Schedule</option>
                      </select>
                      <label class="switch-small">
                        <input type="checkbox" formControlName="value">
                        <span class="slider-small round"></span>
                      </label>
                      <span class="rule-val-label">{{ rule.get('value')?.value ? 'Permit' : 'Deny' }}</span>
                      <button type="button" class="btn-icon delete-btn" (click)="removeRule($index)">×</button>
                    </div>
                    
                    <div class="rule-config" [formGroupName]="'config'">
                      <!-- ... rule inputs same as before ... -->
                      @if (rule.get('type')?.value === 'user_list') {
                        <div class="form-group no-margin">
                          <label class="inner-label">User IDs</label>
                          <input formControlName="userIds" placeholder="e.g. user-1, user-2">
                        </div>
                      }
                      @if (rule.get('type')?.value === 'attribute') {
                        <div class="attr-row">
                          <div class="inner-field">
                            <label class="inner-label">Key</label>
                            <input formControlName="attributeKey" placeholder="plan">
                          </div>
                          <div class="inner-field">
                            <label class="inner-label">Op</label>
                            <select formControlName="attributeOp">
                              <option value="eq">==</option>
                              <option value="neq">!=</option>
                              <option value="contains">contains</option>
                              <option value="gt">></option>
                              <option value="lt"><</option>
                            </select>
                          </div>
                          <div class="inner-field">
                            <label class="inner-label">Value</label>
                            <input formControlName="attributeValue" placeholder="premium">
                          </div>
                        </div>
                      }
                      @if (rule.get('type')?.value === 'percentage') {
                        <div class="range-group">
                          <label class="inner-label">Traffic %</label>
                          <div class="slider-val-row">
                            <input type="range" formControlName="percentage" min="0" max="100" step="0.1">
                            <span class="val-text">{{ rule.get('config.percentage')?.value }}%</span>
                          </div>
                        </div>
                      }
                      @if (rule.get('type')?.value === 'geography') {
                        <div class="geo-grid">
                          <div class="inner-field"><label class="inner-label">Countries</label><input formControlName="countries"></div>
                          <div class="inner-field"><label class="inner-label">Cities</label><input formControlName="cities"></div>
                        </div>
                      }
                      @if (rule.get('type')?.value === 'gradual') {
                        <div class="gradual-grid">
                          <div class="inner-field"><label class="inner-label">Start %</label><input type="number" formControlName="startPercent"></div>
                          <div class="inner-field"><label class="inner-label">End %</label><input type="number" formControlName="endPercent"></div>
                        </div>
                      }
                      @if (rule.get('type')?.value === 'schedule') {
                        <div class="schedule-grid">
                          <div class="inner-field"><label class="inner-label">From</label><input type="datetime-local" formControlName="enableAt"></div>
                          <div class="inner-field"><label class="inner-label">To</label><input type="datetime-local" formControlName="disableAt"></div>
                        </div>
                      }
                    </div>
                  </div>
                } @empty {
                  <div class="empty-rules">
                    <p>No targeting rules defined.</p>
                  </div>
                }
                <button type="button" class="btn btn-outline btn-block mt-md" (click)="addRule()">+ Add Targeting Rule</button>
              </div>

              <!-- Fallthrough (The "Else") -->
              <div class="fallthrough-box">
                <div class="decision-arrow">↓</div>
                <div class="fallthrough-content">
                  <div class="control-item">
                    <label class="switch">
                      <input type="checkbox" formControlName="fallthroughValue">
                      <span class="slider round"></span>
                    </label>
                    <div class="label-group">
                      <span class="font-medium">Fallthrough Value (Default Rule)</span>
                      <span class="text-xs text-muted">Value returned if the flag is ENABLED but NO rules above match.</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="form-actions">
            <button type="submit" class="btn btn-primary" [disabled]="saving()">
              {{ saving() ? 'Saving...' : (isNew ? 'Create Flag' : 'Save Changes') }}
            </button>
            @if (!isNew) {
              <button type="button" class="btn btn-danger" (click)="delete()">Delete</button>
            }
          </div>
          @if (saveSuccess()) {
            <div class="save-success">✓ Flag saved successfully</div>
          }
        </form>
      </div>

      <div class="test-section">
        <div class="test-header">
          <h2>Try it out</h2>
          <button class="btn-text" (click)="refreshTestJSON()">↺ Sync from rules</button>
        </div>
        <div class="card">
          <p class="text-muted">Simulate evaluation with custom context.</p>
          <div class="form-group">
            <label>User ID</label>
            <input [(ngModel)]="testContext.userId" placeholder="e.g. user-123">
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Country</label>
              <input [(ngModel)]="testContext.country" placeholder="e.g. US">
            </div>
            <div class="form-group">
              <label>State</label>
              <input [(ngModel)]="testContext.state" placeholder="e.g. NY">
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>City</label>
              <input [(ngModel)]="testContext.city" placeholder="e.g. New York">
            </div>
            <div class="form-group">
              <label>Zip Code</label>
              <input [(ngModel)]="testContext.zipCode" placeholder="e.g. 10001">
            </div>
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
      gap: var(--space-md);
      margin-top: var(--space-md);
    }
    .switch-label-group {
      display: flex;
      flex-direction: column;
      line-height: 1.2;
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
    .geo-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
    .gradual-grid { display: grid; grid-template-columns: 80px 80px 1fr 1fr; gap: 8px; }
    .schedule-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
    .inner-field { display: flex; flex-direction: column; gap: 4px; }
    .inner-label { font-size: 0.65rem; text-transform: uppercase; font-weight: 700; color: #94a3b8; margin: 0; }
    .no-margin { margin: 0; }
    .range-group { display: flex; flex-direction: column; gap: 4px; }
    .slider-val-row { display: flex; align-items: center; gap: 12px; }
    .val-text { font-size: 0.875rem; font-weight: 600; min-width: 45px; color: var(--primary); }

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
    /* Section Styles */
    .section-title {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: var(--space-lg);
      border-bottom: 1px solid #f1f5f9;
      padding-bottom: 12px;
    }
    .section-title h3 { font-size: 1rem; margin: 0; }
    .step-badge {
      background: #e2e8f0;
      color: #475569;
      width: 24px;
      height: 24px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 50%;
      font-size: 0.75rem;
      font-weight: 700;
    }

    .highlight-card { border: 1px solid #bfdbfe; background: #eff6ff; }
    .card-off { border: 1px solid #fee2e2; background: #fff1f2; }
    .dimmed { opacity: 0.6; filter: grayscale(0.5); pointer-events: none; }
    .dimmed .pill-muted { opacity: 1; }

    .control-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 2rem; }
    .control-item { display: flex; align-items: center; gap: var(--space-md); }
    .label-group { display: flex; flex-direction: column; line-height: 1.3; }
    .control-label { font-weight: 800; font-size: 1.1rem; color: var(--text-main); letter-spacing: 0.05em; }

    .rules-container { padding-left: 36px; border-left: 2px dashed #e2e8f0; margin-left: 12px; }
    .empty-rules { padding: var(--space-md); background: #f8fafc; border-radius: 6px; text-align: center; color: var(--text-muted); font-size: 0.875rem; }
    
    .fallthrough-box { margin-top: 2rem; position: relative; }
    .decision-arrow {
      position: absolute;
      left: -48px;
      top: 50%;
      transform: translateY(-50%);
      font-size: 1.5rem;
      color: #cbd5e1;
      font-weight: bold;
    }
    .fallthrough-content {
      background: #f8fafc;
      padding: var(--space-md);
      border-radius: 8px;
      border: 1px solid var(--border);
    }

    .mb-lg { margin-bottom: var(--space-lg); }
    .ml-auto { margin-left: auto; }
    .mt-md { margin-top: var(--space-md); }
    .btn-block { width: 100%; }
    .btn-outline { background: white; border: 1px solid var(--primary); color: var(--primary); }
    .btn-outline:hover { background: #f0f7ff; }

    /* Switch Styling */
    .switch-large { position: relative; display: inline-block; width: 60px; height: 32px; }
    .switch-large input { opacity: 0; width: 0; height: 0; }
    .slider-large { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #cbd5e1; transition: .4s; }
    .slider-large:before { position: absolute; content: ""; height: 24px; width: 24px; left: 4px; bottom: 4px; background-color: white; transition: .4s; }
    input:checked + .slider-large { background-color: var(--success); }
    input:checked + .slider-large:before { transform: translateX(28px); }
    .slider-large.round { border-radius: 34px; }
    .slider-large.round:before { border-radius: 50%; }
  `]
})
export class FlagDetailComponent implements OnInit {
  isNew = true;
  flagId: string | null = null;
  flagForm: FormGroup;
  saving = signal(false);
  saveSuccess = signal(false);
  
  testContext: any = { userId: '', country: '', state: '', city: '', zipCode: '', attributes: {} };
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
      offValue: [false],
      fallthroughValue: [false],
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
        this.refreshTestJSON();
      });
    }
  }

  get rules() { return this.flagForm.get('rules') as FormArray; }

  refreshTestJSON() {
    const rules = this.flagForm.value.rules || [];
    const context: any = { userId: '', country: '', state: '', city: '', zipCode: '', attributes: {} };
    
    rules.forEach((r: any) => {
      const c = r.config;
      if (r.type === 'user_list' && c.userIds) {
        const ids = (c.userIds || '').split(',').map((s: string) => s.trim()).filter((s: string) => s);
        if (ids.length > 0 && !context.userId) context.userId = ids[0];
      }
      if (r.type === 'geography') {
        const split = (val: string) => (val || '').split(',').map((s: string) => s.trim()).filter((s: string) => s);
        const countries = split(c.countries);
        const states = split(c.states);
        const cities = split(c.cities);
        const zips = split(c.zipCodes);

        if (countries.length > 0 && !context.country) context.country = countries[0];
        if (states.length > 0 && !context.state) context.state = states[0];
        if (cities.length > 0 && !context.city) context.city = cities[0];
        if (zips.length > 0 && !context.zipCode) context.zipCode = zips[0];
      }
      if (r.type === 'attribute' && c.attributeKey) {
        context.attributes[c.attributeKey] = c.attributeValue || 'test-value';
      }
    });

    this.testContext.userId = context.userId || 'user-123';
    this.testContext.country = context.country || 'US';
    this.testContext.state = context.state || '';
    this.testContext.city = context.city || '';
    this.testContext.zipCode = context.zipCode || '';
    this.testAttributes = JSON.stringify(context.attributes, null, 2);
  }

  addRule() {
    const ruleForm = this.fb.group({
      type: ['user_list'],
      value: [true],
      config: this.fb.group({
        userIds: [''],
        attributeKey: [''],
        attributeOp: ['eq'],
        attributeValue: [''],
        percentage: [50],
        countries: [''],
        cities: [''],
        states: [''],
        zipCodes: [''],
        startPercent: [0],
        endPercent: [100],
        startAt: [''],
        endAt: [''],
        enableAt: [''],
        disableAt: ['']
      })
    });
    this.rules.push(ruleForm);
  }

  removeRule(index: number) { this.rules.removeAt(index); }

  private formatDateForInput(dateStr: string | undefined): string {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return '';
    return date.toISOString().slice(0, 16); // YYYY-MM-DDTHH:mm
  }

  patchFlag(flag: Flag) {
    this.flagForm.patchValue({
      name: flag.name,
      key: flag.key,
      description: flag.description,
      enabled: flag.enabled,
      offValue: flag.offValue,
      fallthroughValue: flag.fallthroughValue,
      ruleMatchStrategy: flag.ruleMatchStrategy
    });
    
    flag.rules?.forEach(r => {
      const config = { ...r.config };
      
      const ruleForm = this.fb.group({
        type: [r.type],
        value: [r.value],
        config: this.fb.group({
          userIds: [(config.userIds || []).join(', ')],
          attributeKey: [config.attributeKey || ''],
          attributeOp: [config.attributeOp || 'eq'],
          attributeValue: [config.attributeValue || ''],
          percentage: [config.percentage || 50],
          countries: [(config.countries || []).join(', ')],
          cities: [(config.cities || []).join(', ')],
          states: [(config.states || []).join(', ')],
          zipCodes: [(config.zipCodes || []).join(', ')],
          startPercent: [config.startPercent || 0],
          endPercent: [config.endPercent || 100],
          startAt: [this.formatDateForInput(config.startAt)],
          endAt: [this.formatDateForInput(config.endAt)],
          enableAt: [this.formatDateForInput(config.enableAt)],
          disableAt: [this.formatDateForInput(config.disableAt)]
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
      const c = r.config;

      const split = (s: string) => (s || '').split(',').map(v => v.trim()).filter(v => v);
      const toIso = (s: string) => s ? new Date(s).toISOString() : undefined;

      switch (r.type) {
        case 'user_list':
          config.userIds = split(c.userIds);
          break;
        case 'attribute':
          config.attributeKey = c.attributeKey;
          config.attributeOp = c.attributeOp;
          config.attributeValue = c.attributeValue;
          break;
        case 'percentage':
          config.percentage = parseFloat(c.percentage);
          break;
        case 'geography':
          config.countries = split(c.countries);
          config.cities = split(c.cities);
          config.states = split(c.states);
          config.zipCodes = split(c.zipCodes);
          break;
        case 'gradual':
          config.startPercent = parseFloat(c.startPercent);
          config.endPercent = parseFloat(c.endPercent);
          config.startAt = toIso(c.startAt);
          config.endAt = toIso(c.endAt);
          break;
        case 'schedule':
          config.enableAt = toIso(c.enableAt);
          config.disableAt = toIso(c.disableAt);
          break;
      }
      return { type: r.type, value: r.value, config };
    });

    const flagData = { ...formVal, rules };

    const obs = this.isNew 
      ? this.flagService.createFlag(flagData) 
      : this.flagService.updateFlag(this.flagId!, flagData);

    obs.subscribe({
      next: (res: any) => {
        this.saving.set(false);
        this.saveSuccess.set(true);
        setTimeout(() => this.saveSuccess.set(false), 3000);
        
        if (this.isNew && res.id) {
          this.isNew = false;
          this.flagId = res.id;
          // Update URL without reloading to reflect existing ID
          this.router.navigate(['/flags', res.id], { replaceUrl: true });
        }
      },
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
