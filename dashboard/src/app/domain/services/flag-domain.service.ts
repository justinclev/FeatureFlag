import { Injectable } from '@angular/core';
import { Flag, Rule, RuleType, RuleConfig } from '../models/flag.domain';

@Injectable({ providedIn: 'root' })
export class FlagDomainService {
  /**
   * Transforms raw form rule data into the domain format required by the API.
   * Handles string-to-array splitting and number parsing.
   */
  mapFormToRule(formRule: any): Rule {
    const c = formRule.config;
    const config: RuleConfig = {};

    const split = (s: string) => (s || '').split(',').map(v => v.trim()).filter(v => v);
    const toIso = (s: string) => s ? new Date(s).toISOString() : undefined;

    switch (formRule.type) {
      case RuleType.UserList: config.userIds = split(c.userIds); break;
      case RuleType.Attribute: 
        config.attributeKey = c.attributeKey;
        config.attributeOp = c.attributeOp;
        config.attributeValue = c.attributeValue;
        break;
      case RuleType.Percentage: config.percentage = parseFloat(c.percentage); break;
      case RuleType.Geography:
        config.countries = split(c.countries);
        config.cities = split(c.cities);
        config.states = split(c.states);
        config.zipCodes = split(c.zipCodes);
        break;
      case RuleType.Gradual:
        config.startPercent = parseFloat(c.startPercent);
        config.endPercent = parseFloat(c.endPercent);
        config.startAt = toIso(c.startAt);
        config.endAt = toIso(c.endAt);
        break;
      case RuleType.Schedule:
        config.enableAt = toIso(c.enableAt);
        config.disableAt = toIso(c.disableAt);
        break;
    }

    return {
      description: formRule.description,
      type: formRule.type,
      value: formRule.value,
      config
    };
  }

  /**
   * Transforms domain rule data into the flat structure needed by Reactive Forms.
   */
  mapRuleToForm(rule: Rule): any {
    const config = { ...rule.config };
    return {
      description: rule.description || '',
      type: rule.type,
      value: rule.value,
      config: {
        userIds: (config.userIds || []).join(', '),
        attributeKey: config.attributeKey || '',
        attributeOp: config.attributeOp || 'eq',
        attributeValue: config.attributeValue || '',
        percentage: config.percentage || 50,
        countries: (config.countries || []).join(', '),
        cities: (config.cities || []).join(', '),
        states: (config.states || []).join(', '),
        zipCodes: (config.zipCodes || []).join(', '),
        startPercent: config.startPercent || 0,
        endPercent: config.endPercent || 100,
        startAt: this.formatDateForInput(config.startAt),
        endAt: this.formatDateForInput(config.endAt),
        enableAt: this.formatDateForInput(config.enableAt),
        disableAt: this.formatDateForInput(config.disableAt)
      }
    };
  }

  private formatDateForInput(dateStr: string | undefined): string {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return '';
    return date.toISOString().slice(0, 16);
  }
}
