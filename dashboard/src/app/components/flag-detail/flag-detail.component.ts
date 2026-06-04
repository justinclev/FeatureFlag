import { Component, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { FlagService } from '../../services/flag.service';
import { Flag, HistoryEntry, Rule, RuleType, MatchStrategy } from '../../domain/models/flag.domain';
import { IdentitySafetyEditorComponent } from './components/identity-safety-editor/identity-safety-editor.component';
import { TargetingRulesEditorComponent } from './components/targeting-rules-editor/targeting-rules-editor.component';
import { EvaluationPlaygroundComponent } from './components/evaluation-playground/evaluation-playground.component';

import { FlagDomainService } from '../../domain/services/flag-domain.service';
import { allStrategyValidator } from '../../domain/validators/flag.validators';

@Component({
  selector: 'app-flag-detail',
  standalone: true,
  imports: [
    CommonModule,
    DatePipe,
    RouterLink,
    ReactiveFormsModule,
    IdentitySafetyEditorComponent,
    TargetingRulesEditorComponent,
    EvaluationPlaygroundComponent
  ],
  templateUrl: './flag-detail.component.html',
  styleUrl: './flag-detail.component.css'
})
export class FlagDetailComponent implements OnInit {
  isNew = true;
  flagId: string | null = null;
  flagForm: FormGroup;
  loading = signal(false);
  saving = signal(false);
  saveSuccess = signal(false);
  error = signal<string | null>(null);
  loadError = signal<string | null>(null);
  flagHistory = signal<HistoryEntry[]>([]);

  constructor(
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private flagService: FlagService,
    private domainService: FlagDomainService
  ) {
    this.flagForm = this.fb.group({
      name: ['', Validators.required],
      key: ['', [Validators.required, Validators.pattern(/^[a-z0-9._-]+$/)]],
      description: [''],
      enabled: [true],
      offValue: [false],
      fallthroughValue: [false],
      ruleMatchStrategy: [MatchStrategy.Any],
      rules: this.fb.array([])
    }, { validators: [allStrategyValidator] });
  }

  ngOnInit() {
    this.flagId = this.route.snapshot.paramMap.get('id');
    if (this.flagId && this.flagId !== 'new') {
      this.isNew = false;
      this.loading.set(true);
      this.flagService.getFlag(this.flagId).subscribe({
        next: (flag) => {
          this.patchFlag(flag);
          this.flagHistory.set([...(flag.history ?? [])].reverse());
          this.loading.set(false);
        },
        error: (err) => {
          this.loadError.set(err.status === 404 ? 'Flag not found.' : 'Failed to load flag.');
          this.loading.set(false);
        }
      });
    }
  }

  addRule() {
    const emptyRule: Rule = {
      description: '',
      type: RuleType.UserList,
      value: true,
      config: {}
    };
    const ruleForm = this.fb.group(this.domainService.mapRuleToForm(emptyRule));
    const rules = this.flagForm.get('rules') as any;
    rules.push(ruleForm);
  }

  removeRule(index: number) {
    const rules = this.flagForm.get('rules') as any;
    rules.removeAt(index);
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

    // Key is immutable after creation — lock it so it can't be changed from the UI
    this.flagForm.get('key')?.disable();

    const rulesArray = this.flagForm.get('rules') as any;
    rulesArray.clear();
    flag.rules?.forEach(r => {
      rulesArray.push(this.fb.group(this.domainService.mapRuleToForm(r)));
    });
  }

  save() {
    if (this.flagForm.invalid) return;
    this.saving.set(true);
    this.error.set(null);
    
    const formVal = this.flagForm.getRawValue();
    const rules = formVal.rules.map((r: any) => this.domainService.mapFormToRule(r));

    const flagData = { ...formVal, rules };
    const obs = this.isNew 
      ? this.flagService.createFlag(flagData) 
      : this.flagService.updateFlag(this.flagId!, flagData);

    obs.subscribe({
      next: (res: any) => {
        this.saving.set(false);
        this.saveSuccess.set(true);
        this.error.set(null);
        this.flagHistory.set([...(res.history ?? [])].reverse());
        setTimeout(() => this.saveSuccess.set(false), 3000);
        if (this.isNew && res.id) {
          this.isNew = false;
          this.flagId = res.id;
          this.router.navigate(['/flags', res.id], { replaceUrl: true });
        }
      },
      error: (err) => {
        this.error.set(err.error?.error || 'Failed to save flag');
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
}
