// Mirrors the JSON served by tracker.exe (cmd/daemon/analyze.go, server.go).

export type Kind = "productive" | "distraction" | "neutral";

export interface NamedSecs {
  name: string;
  seconds: number;
}

export interface CategoryStat {
  name: string;
  seconds: number;
  kind: Kind;
}

export interface SubjectStat {
  name: string;
  seconds: number;
  topics: NamedSecs[];
}

export interface AppStat {
  name: string;
  exe: string;
  seconds: number;
  productive: number;
  distraction: number;
}

export interface SiteStat {
  site: string;
  seconds: number;
  kind: Kind;
}

export interface ActivityStat {
  label: string;
  app: string;
  site: string;
  category: string;
  subject: string;
  topic: string;
  kind: Kind;
  seconds: number;
}

export interface TimelineBlock {
  start: number;
  end: number;
  kind: Kind;
  category: string;
  label: string;
  app: string;
}

export interface DeepWorkBlock {
  start: number;
  end: number;
  seconds: number;
  category: string;
  label: string;
}

export interface DayReport {
  total: number;
  productive: number;
  distraction: number;
  other: number;
  focusScore: number | null;
  byCategory: CategoryStat[];
  bySubject: SubjectStat[];
  byApp: AppStat[];
  bySite: SiteStat[];
  activities: ActivityStat[];
  timeline: TimelineBlock[];
  deepWork: DeepWorkBlock[];
  firstTs: number;
  lastTs: number;
}

export interface WeekDay {
  date: string;
  label: string;
  productive: number;
  distraction: number;
  other: number;
}

export interface AIStatus {
  enabled: boolean;
  configured: boolean;
  running: boolean;
  lastRun: number;
  lastError: string;
  labeled: number;
  pending: number;
}

export interface Stats {
  date: string;
  isToday: boolean;
  paused: boolean;
  ai: AIStatus;
  day: DayReport;
  week: WeekDay[];
}

export interface Settings {
  ai: {
    enabled: boolean;
    provider: string;
    baseURL: string;
    model: string;
    hasKey: boolean;
    keyHint: string;
  };
  startWithWindows: boolean;
  dataDir: string;
  rulesPath: string;
  dashboardURL: string;
}
