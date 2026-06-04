import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormGroup, FormArray } from '@angular/forms';

@Component({
  selector: 'app-targeting-rules-editor',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './targeting-rules-editor.component.html',
  styleUrl: './targeting-rules-editor.component.css'
})
export class TargetingRulesEditorComponent {
  @Input({ required: true }) form!: FormGroup;
  @Input({ required: true }) enabled = true;
  @Output() addRule = new EventEmitter<void>();
  @Output() removeRule = new EventEmitter<number>();

  get rules() { return this.form.get('rules') as FormArray; }
}
