import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { IdentitySafetyEditorComponent } from './identity-safety-editor.component';

describe('IdentitySafetyEditorComponent', () => {
  let component: IdentitySafetyEditorComponent;
  let fixture: ComponentFixture<IdentitySafetyEditorComponent>;
  let fb: FormBuilder;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ReactiveFormsModule, IdentitySafetyEditorComponent]
    }).compileComponents();

    fb = TestBed.inject(FormBuilder);
    fixture = TestBed.createComponent(IdentitySafetyEditorComponent);
    component = fixture.componentInstance;
    component.form = fb.group({
      name: [''],
      description: [''],
      key: [''],
      ruleMatchStrategy: ['any'],
      enabled: [true],
      offValue: [false]
    });
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
