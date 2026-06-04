import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { FlagService } from './flag.service';
import { MatchStrategy } from '../domain/models/flag.domain';

describe('FlagService', () => {
  let service: FlagService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        FlagService,
        provideHttpClient(),
        provideHttpClientTesting()
      ]
    });
    service = TestBed.inject(FlagService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should load flags into state', () => {
    const mockFlags = [{ id: '1', key: 'test', name: 'Test', ruleMatchStrategy: MatchStrategy.Any, rules: [], enabled: true, offValue: false, fallthroughValue: false, description: '' }];
    
    service.loadFlags().subscribe();
    
    const req = httpMock.expectOne('http://127.0.0.1:8081/api/flags');
    req.flush(mockFlags);

    expect(service.flags()).toEqual(mockFlags);
    expect(service.activeCount()).toBe(1);
  });
});
