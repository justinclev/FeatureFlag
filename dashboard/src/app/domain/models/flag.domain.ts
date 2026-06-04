export enum RuleType {
  UserList = 'user_list',
  Attribute = 'attribute',
  Percentage = 'percentage',
  Geography = 'geography',
  Gradual = 'gradual',
  Schedule = 'schedule'
}

export enum MatchStrategy {
  Any = 'any',
  All = 'all'
}

export interface RuleConfig {
  userIds?: string[];
  attributeKey?: string;
  attributeOp?: string;
  attributeValue?: string;
  percentage?: number;
  countries?: string[];
  states?: string[];
  cities?: string[];
  zipCodes?: string[];
  startPercent?: number;
  endPercent?: number;
  startAt?: string;
  endAt?: string;
  enableAt?: string;
  disableAt?: string;
}

export interface Rule {
  id?: string;
  description?: string;
  type: RuleType;
  config: RuleConfig;
  value: boolean;
}

export interface HistoryEntry {
  changedAt: string;
  changedBy: string;
  summary: string;
}

export interface Flag {
  id?: string;
  key: string;
  name: string;
  description: string;
  enabled: boolean;
  offValue: boolean;
  fallthroughValue: boolean;
  ruleMatchStrategy: MatchStrategy;
  rules: Rule[];
  createdAt?: string;
  createdBy?: string;
  updatedAt?: string;
  updatedBy?: string;
  history?: HistoryEntry[];
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
