export interface Rule {
  id?: string;
  type: string;
  config: any;
  value: boolean;
}

export interface Flag {
  id?: string;
  key: string;
  name: string;
  description: string;
  enabled: boolean;
  offValue: boolean;
  fallthroughValue: boolean;
  ruleMatchStrategy: 'any' | 'all';
  rules: Rule[];
  createdAt?: string;
  createdBy?: string;
  updatedAt?: string;
  updatedBy?: string;
}

export interface EvaluationContext {
  userId?: string;
  country?: string;
  state?: string;
  city?: string;
  zipCode?: string;
  attributes?: Record<string, any>;
}

export interface EvaluationResult {
  enabled: boolean;
  reason: string;
}
