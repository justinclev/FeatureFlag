import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { TargetingRulesEditorComponent } from './targeting-rules-editor.component';

describe('TargetingRulesEditorComponent', () => {
  let component: TargetingRulesEditorComponent;
  let fixture: ComponentFixture<TargetingRulesEditorComponent>;
  let fb: FormBuilder;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ReactiveFormsModule, TargetingRulesEditorComponent]
    }).compileComponents();

    fb = TestBed.inject(FormBuilder);
    fixture = TestBed.createComponent(TargetingRulesEditorComponent);
    component = fixture.componentInstance;
    component.form = fb.group({
      fallthroughValue: [false],
      rules: fb.array([])
    });
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
