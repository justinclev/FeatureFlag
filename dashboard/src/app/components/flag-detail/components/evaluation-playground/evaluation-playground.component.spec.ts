import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { EvaluationPlaygroundComponent } from './evaluation-playground.component';

describe('EvaluationPlaygroundComponent', () => {
  let component: EvaluationPlaygroundComponent;
  let fixture: ComponentFixture<EvaluationPlaygroundComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EvaluationPlaygroundComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting()
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(EvaluationPlaygroundComponent);
    component = fixture.componentInstance;
    component.flagKey.set('test-key');
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
