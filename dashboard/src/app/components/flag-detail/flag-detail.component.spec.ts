import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter, ActivatedRoute } from '@angular/router';
import { of, throwError } from 'rxjs';
import { FlagDetailComponent } from './flag-detail.component';
import { FlagService } from '../../services/flag.service';
import { Flag, MatchStrategy, RuleType } from '../../domain/models/flag.domain';

const mockFlag: Flag = {
  id: 'abc123',
  key: 'my-flag',
  name: 'My Flag',
  description: 'Test flag',
  enabled: true,
  offValue: false,
  fallthroughValue: true,
  ruleMatchStrategy: MatchStrategy.Any,
  rules: [],
  history: [
    { changedAt: '2026-01-01T00:00:00Z', changedBy: 'alice', summary: 'Flag created' }
  ]
};

function makeRoute(id: string) {
  return { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => id } } } };
}

describe('FlagDetailComponent', () => {
  let component: FlagDetailComponent;
  let fixture: ComponentFixture<FlagDetailComponent>;
  let flagService: FlagService;

  async function setup(routeId = 'new') {
    await TestBed.configureTestingModule({
      imports: [FlagDetailComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        makeRoute(routeId)
      ]
    }).compileComponents();

    flagService = TestBed.inject(FlagService);
    fixture = TestBed.createComponent(FlagDetailComponent);
    component = fixture.componentInstance;
  }

  describe('new flag mode', () => {
    beforeEach(async () => setup('new'));

    it('creates in new mode with isNew=true', () => {
      fixture.detectChanges();
      expect(component.isNew).toBeTrue();
      expect(component.flagId).toBe('new');
    });

    it('form starts valid enough to show key field enabled', () => {
      fixture.detectChanges();
      expect(component.flagForm.get('key')?.disabled).toBeFalse();
    });

    it('form is invalid when name and key are empty', () => {
      fixture.detectChanges();
      expect(component.flagForm.invalid).toBeTrue();
    });

    it('save() does nothing when form is invalid', () => {
      fixture.detectChanges();
      spyOn(flagService, 'createFlag').and.returnValue(of(mockFlag));
      component.save();
      expect(flagService.createFlag).not.toHaveBeenCalled();
    });

    it('key validator rejects uppercase', () => {
      fixture.detectChanges();
      const key = component.flagForm.get('key');
      key?.setValue('MyFlag');
      key?.markAsTouched();
      expect(key?.errors?.['pattern']).toBeTruthy();
    });

    it('key validator accepts dots and underscores', () => {
      fixture.detectChanges();
      const key = component.flagForm.get('key');
      key?.setValue('my_flag.v2');
      expect(key?.errors).toBeNull();
    });
  });

  describe('edit flag mode', () => {
    beforeEach(async () => setup('abc123'));

    it('loads flag and patches form on success', fakeAsync(() => {
      spyOn(flagService, 'getFlag').and.returnValue(of(mockFlag));
      fixture.detectChanges();
      tick();
      expect(component.isNew).toBeFalse();
      expect(component.flagForm.get('name')?.value).toBe('My Flag');
      expect(component.loading()).toBeFalse();
    }));

    it('disables key field after load', fakeAsync(() => {
      spyOn(flagService, 'getFlag').and.returnValue(of(mockFlag));
      fixture.detectChanges();
      tick();
      expect(component.flagForm.get('key')?.disabled).toBeTrue();
    }));

    it('populates flagHistory in reverse chronological order', fakeAsync(() => {
      const flagWithHistory: Flag = {
        ...mockFlag,
        history: [
          { changedAt: '2026-01-01T00:00:00Z', changedBy: 'alice', summary: 'Flag created' },
          { changedAt: '2026-02-01T00:00:00Z', changedBy: 'bob', summary: 'Updated: enabled' }
        ]
      };
      spyOn(flagService, 'getFlag').and.returnValue(of(flagWithHistory));
      fixture.detectChanges();
      tick();
      const history = component.flagHistory();
      expect(history[0].summary).toBe('Updated: enabled');
      expect(history[1].summary).toBe('Flag created');
    }));

    it('sets loadError and clears loading on 404', fakeAsync(() => {
      spyOn(flagService, 'getFlag').and.returnValue(throwError(() => ({ status: 404 })));
      fixture.detectChanges();
      tick();
      expect(component.loadError()).toBe('Flag not found.');
      expect(component.loading()).toBeFalse();
    }));

    it('sets generic loadError on non-404 failure', fakeAsync(() => {
      spyOn(flagService, 'getFlag').and.returnValue(throwError(() => ({ status: 500 })));
      fixture.detectChanges();
      tick();
      expect(component.loadError()).toBe('Failed to load flag.');
    }));
  });

  describe('addRule / removeRule', () => {
    beforeEach(async () => setup('new'));

    it('addRule appends a rule to the form array', () => {
      fixture.detectChanges();
      expect(component.flagForm.get('rules')?.value.length).toBe(0);
      component.addRule();
      expect(component.flagForm.get('rules')?.value.length).toBe(1);
    });

    it('removeRule removes the rule at the given index', () => {
      fixture.detectChanges();
      component.addRule();
      component.addRule();
      component.removeRule(0);
      expect(component.flagForm.get('rules')?.value.length).toBe(1);
    });
  });
});
