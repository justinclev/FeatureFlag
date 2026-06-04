import { TestBed } from '@angular/core/testing';
import { FlagDomainService } from './flag-domain.service';
import { RuleType } from '../models/flag.domain';

describe('FlagDomainService', () => {
  let service: FlagDomainService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(FlagDomainService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should map form rule to domain rule (UserList)', () => {
    const formRule = {
      description: 'test rule',
      type: RuleType.UserList,
      value: true,
      config: { userIds: 'u1, u2' }
    };
    const domainRule = service.mapFormToRule(formRule);
    expect(domainRule.config.userIds).toEqual(['u1', 'u2']);
    expect(domainRule.value).toBe(true);
  });

  it('should map form rule to domain rule (Percentage)', () => {
    const formRule = { type: RuleType.Percentage, value: true, config: { percentage: '25.5' } };
    const domainRule = service.mapFormToRule(formRule);
    expect(domainRule.config.percentage).toBe(25.5);
  });

  it('should map form rule to domain rule (Gradual)', () => {
    const startInput = '2026-05-24T10:00';
    const formRule = {
      type: RuleType.Gradual,
      value: true,
      config: { startPercent: '10', endPercent: '90', startAt: startInput, endAt: '2026-05-31T10:00' }
    };
    const domainRule = service.mapFormToRule(formRule);
    expect(domainRule.config.startPercent).toBe(10);
    // ISO output must round-trip back to the same local time
    expect(domainRule.config.startAt).toBe(new Date(startInput).toISOString());
  });

  it('should map form rule to domain rule (Schedule)', () => {
    const enableInput = '2026-05-24T10:00';
    const formRule = {
      type: RuleType.Schedule,
      value: true,
      config: { enableAt: enableInput, disableAt: '2026-05-25T10:00' }
    };
    const domainRule = service.mapFormToRule(formRule);
    expect(domainRule.config.enableAt).toBe(new Date(enableInput).toISOString());
  });

  it('should map domain rule to form rule (Schedule)', () => {
    const domainRule = {
      type: RuleType.Schedule,
      value: true,
      config: { enableAt: '2026-05-24T10:00:00Z', disableAt: '2026-05-25T10:00:00Z' }
    };
    const formRule = service.mapRuleToForm(domainRule);
    expect(formRule.config.enableAt).toBe('2026-05-24T10:00');
  });
});
