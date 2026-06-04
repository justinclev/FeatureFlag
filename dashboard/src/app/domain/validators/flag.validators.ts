import { AbstractControl, ValidationErrors, ValidatorFn } from '@angular/forms';
import { MatchStrategy } from '../models/flag.domain';

export const allStrategyValidator: ValidatorFn = (control: AbstractControl): ValidationErrors | null => {
  const strategy = control.get('ruleMatchStrategy')?.value;
  const rules = control.get('rules')?.value || [];

  if (strategy === MatchStrategy.All && rules.length > 1) {
    const firstValue = rules[0].value;
    const allSame = rules.every((r: any) => r.value === firstValue);
    if (!allSame) {
      return { inconsistentRules: true };
    }
  }

  return null;
};
