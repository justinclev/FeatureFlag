import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { FlagListComponent } from './flag-list.component';
import { FlagService } from '../../services/flag.service';
import { Flag, MatchStrategy } from '../../domain/models/flag.domain';
import { environment } from '../../../environments/environment';

const mockFlags: Flag[] = [
  {
    id: '1', key: 'flag-a', name: 'Flag A', description: '', enabled: true,
    offValue: false, fallthroughValue: false, ruleMatchStrategy: MatchStrategy.Any, rules: []
  },
  {
    id: '2', key: 'flag-b', name: 'Flag B', description: '', enabled: false,
    offValue: false, fallthroughValue: false, ruleMatchStrategy: MatchStrategy.Any, rules: []
  }
];

describe('FlagListComponent', () => {
  let component: FlagListComponent;
  let fixture: ComponentFixture<FlagListComponent>;
  let flagService: FlagService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FlagListComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([])
      ]
    }).compileComponents();

    flagService = TestBed.inject(FlagService);
  });

  it('calls loadFlags on init', () => {
    spyOn(flagService, 'loadFlags').and.returnValue(of([]));
    fixture = TestBed.createComponent(FlagListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    expect(flagService.loadFlags).toHaveBeenCalled();
  });

  it('reflects activeCount from service after flags load', fakeAsync(() => {
    fixture = TestBed.createComponent(FlagListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    const httpMock = TestBed.inject(HttpTestingController);
    const req = httpMock.expectOne(`${environment.apiUrl}/flags`);
    req.flush(mockFlags);
    tick();
    httpMock.verify();

    // 1 of 2 flags is enabled
    expect(flagService.activeCount()).toBe(1);
  }));

  it('creates component', () => {
    spyOn(flagService, 'loadFlags').and.returnValue(of([]));
    fixture = TestBed.createComponent(FlagListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });
});
