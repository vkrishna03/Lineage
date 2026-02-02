/**
 * Type definitions for Lineage SDK
 */

// Enums
export type ActorType = 'human' | 'llm' | 'agent' | 'service' | 'tool';
export type Intent = 'exploration' | 'suggestion' | 'assertion' | 'decision' | 'execution';
export type CorrectionType = 'supersede' | 'amend' | 'retract';
export type ScoreType = 'confidence' | 'relevance' | 'reliability' | 'agreement';
export type ScoreCategory = 'very_low' | 'low' | 'moderate' | 'high' | 'very_high';
export type ArtifactRole = 'input' | 'output';

// Models
export interface Scope {
  id: string;
  project: string;
  domain?: string;
  environment?: string;
  created_at: string;
}

export interface Actor {
  id: string;
  type: ActorType;
  external_id?: string;
  name?: string;
  metadata?: Record<string, unknown>;
  registered_at: string;
}

export interface EventType {
  id: string;
  name: string;
  version: string;
  description?: string;
  payload_schema?: Record<string, unknown>;
  allowed_intents: string[];
  is_active: boolean;
  created_at: string;
}

export interface Event {
  id: string;
  scope_id: string;
  actor_id: string;
  event_type_id: string;
  scope_sequence: number;
  intent: Intent;
  reason?: string;
  correction_type?: CorrectionType;
  corrects_event_id?: string;
  observed_at?: string;
  decided_at?: string;
  ingested_at: string;
  prev_event_hash?: string;
  event_hash: string;
  payload: Record<string, unknown>;
}

export interface Artifact {
  id: string;
  scope_id: string;
  content_hash: string;
  content_type: string;
  uri?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface Score {
  id: string;
  event_id: string;
  type: ScoreType;
  value: number;
  category: ScoreCategory;
  scored_by?: string;
  reason?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface Lineage {
  event_id: string;
  parents: Event[];
  children: Event[];
}

export interface HealthStatus {
  status: string;
  services: Record<string, string>;
}

// Request/Response types
export interface CreateScopeRequest {
  project: string;
  domain?: string;
  environment?: string;
}

export interface CreateActorRequest {
  type: ActorType;
  external_id?: string;
  name?: string;
  metadata?: Record<string, unknown>;
}

export interface CreateEventTypeRequest {
  name: string;
  version: string;
  description?: string;
  payload_schema?: Record<string, unknown>;
  allowed_intents?: Intent[];
}

export interface CreateEventRequest {
  scope_id: string;
  actor_id: string;
  event_type_id: string;
  intent: Intent;
  payload: Record<string, unknown>;
  reason?: string;
  correction_type?: CorrectionType;
  corrects_event_id?: string;
  observed_at?: string;
  decided_at?: string;
  parent_event_ids?: string[];
  confidence?: number;
  input_artifact_ids?: string[];
  output_artifact_ids?: string[];
}

export interface CreateArtifactRequest {
  scope_id: string;
  content_hash: string;
  content_type: string;
  uri?: string;
  metadata?: Record<string, unknown>;
}

export interface CreateScoreRequest {
  type: ScoreType;
  value: number;
  scored_by?: string;
  reason?: string;
  metadata?: Record<string, unknown>;
}

export interface LinkArtifactRequest {
  artifact_id: string;
  role: ArtifactRole;
}

// SDK configuration
export interface LineageConfig {
  baseUrl?: string;
  timeout?: number;
}

export interface InitConfig {
  project: string;
  baseUrl?: string;
  domain?: string;
  environment?: string;
  actorName?: string;
  actorType?: ActorType;
  autoCreate?: boolean;
  waitTime?: number;
}

export interface EmitOptions {
  confidence?: number;
  actor?: [ActorType, string];
  parent?: Event | string;
  reason?: string;
}

export interface TrackOptions {
  actor?: [ActorType, string];
  confidence?: number;
}
