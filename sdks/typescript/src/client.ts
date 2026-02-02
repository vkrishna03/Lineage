/**
 * Low-level Lineage HTTP client
 */

import type {
  Actor,
  Artifact,
  ArtifactRole,
  CreateActorRequest,
  CreateArtifactRequest,
  CreateEventRequest,
  CreateEventTypeRequest,
  CreateScoreRequest,
  CreateScopeRequest,
  Event,
  EventType,
  HealthStatus,
  Intent,
  Lineage,
  LineageConfig,
  LinkArtifactRequest,
  Score,
  ScoreType,
  Scope,
} from './types';

class LineageError extends Error {
  constructor(
    message: string,
    public statusCode?: number
  ) {
    super(message);
    this.name = 'LineageError';
  }
}

class NotFoundError extends LineageError {
  constructor(resource: string, id: string) {
    super(`${resource} not found: ${id}`, 404);
    this.name = 'NotFoundError';
  }
}

class ValidationError extends LineageError {
  constructor(message: string) {
    super(message, 400);
    this.name = 'ValidationError';
  }
}

class ServerError extends LineageError {
  constructor(message: string = 'Internal server error') {
    super(message, 500);
    this.name = 'ServerError';
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (response.status >= 500) {
    throw new ServerError(await response.text());
  }
  if (response.status === 404) {
    const data = await response.json();
    throw new NotFoundError('Resource', data.error || 'unknown');
  }
  if (response.status === 400) {
    const data = await response.json();
    throw new ValidationError(data.error || 'Validation failed');
  }
  if (response.status >= 400) {
    throw new LineageError(await response.text(), response.status);
  }
  return response.json();
}

function decodeMetadata(value: unknown): Record<string, unknown> | undefined {
  if (value === null || value === undefined) return undefined;
  if (typeof value === 'object') return value as Record<string, unknown>;
  if (typeof value === 'string') {
    try {
      // Try base64 decode (Go API returns []byte as base64)
      const decoded = atob(value);
      return JSON.parse(decoded);
    } catch {
      try {
        return JSON.parse(value);
      } catch {
        return undefined;
      }
    }
  }
  return undefined;
}

function transformEvent(data: Record<string, unknown>): Event {
  return {
    ...data,
    payload: decodeMetadata(data.payload) || {},
    correction_type: data.correction_type && typeof data.correction_type === 'object'
      ? (data.correction_type as { valid: boolean; correction_type: string }).valid
        ? (data.correction_type as { correction_type: string }).correction_type as Event['correction_type']
        : undefined
      : data.correction_type as Event['correction_type'],
    corrects_event_id: data.corrects_event_id && typeof data.corrects_event_id === 'object'
      ? (data.corrects_event_id as { Valid: boolean; UUID: string }).Valid
        ? (data.corrects_event_id as { UUID: string }).UUID
        : undefined
      : data.corrects_event_id as string | undefined,
  } as Event;
}

function transformActor(data: Record<string, unknown>): Actor {
  return {
    ...data,
    metadata: decodeMetadata(data.metadata),
  } as Actor;
}

function transformArtifact(data: Record<string, unknown>): Artifact {
  return {
    ...data,
    metadata: decodeMetadata(data.metadata),
  } as Artifact;
}

function transformScore(data: Record<string, unknown>): Score {
  let value = data.value;
  if (typeof value === 'object' && value !== null) {
    // Handle pgtype.Numeric
    const numeric = value as { Int?: number; Exp?: number; Valid?: boolean };
    if (numeric.Valid && numeric.Int !== undefined) {
      value = numeric.Int * Math.pow(10, numeric.Exp || 0);
    } else {
      value = 0;
    }
  }
  return {
    ...data,
    value: Number(value),
    metadata: decodeMetadata(data.metadata),
    scored_by: data.scored_by && typeof data.scored_by === 'object'
      ? (data.scored_by as { Valid: boolean; UUID: string }).Valid
        ? (data.scored_by as { UUID: string }).UUID
        : undefined
      : data.scored_by as string | undefined,
  } as Score;
}

function transformEventType(data: Record<string, unknown>): EventType {
  return {
    ...data,
    payload_schema: decodeMetadata(data.payload_schema),
  } as EventType;
}

class ScopesResource {
  constructor(private baseUrl: string, private timeout: number) {}

  async create(request: CreateScopeRequest): Promise<Scope> {
    const response = await fetch(`${this.baseUrl}/api/v1/scopes`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal: AbortSignal.timeout(this.timeout),
    });
    return handleResponse<Scope>(response);
  }

  async get(id: string): Promise<Scope> {
    const response = await fetch(`${this.baseUrl}/api/v1/scopes/${id}`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    return handleResponse<Scope>(response);
  }

  async list(): Promise<Scope[]> {
    const response = await fetch(`${this.baseUrl}/api/v1/scopes`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<{ scopes: Scope[] }>(response);
    return data.scopes || [];
  }
}

class ActorsResource {
  constructor(private baseUrl: string, private timeout: number) {}

  async create(request: CreateActorRequest): Promise<Actor> {
    const response = await fetch(`${this.baseUrl}/api/v1/actors`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<Record<string, unknown>>(response);
    return transformActor(data);
  }

  async get(id: string): Promise<Actor> {
    const response = await fetch(`${this.baseUrl}/api/v1/actors/${id}`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<Record<string, unknown>>(response);
    return transformActor(data);
  }

  async list(): Promise<Actor[]> {
    const response = await fetch(`${this.baseUrl}/api/v1/actors`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<{ actors: Record<string, unknown>[] }>(response);
    return (data.actors || []).map(transformActor);
  }
}

class EventTypesResource {
  constructor(private baseUrl: string, private timeout: number) {}

  async create(request: CreateEventTypeRequest): Promise<EventType> {
    const payload = {
      ...request,
      allowed_intents: request.allowed_intents || ['exploration', 'suggestion', 'assertion', 'decision', 'execution'],
    };
    const response = await fetch(`${this.baseUrl}/api/v1/event-types`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<Record<string, unknown>>(response);
    return transformEventType(data);
  }

  async get(id: string): Promise<EventType> {
    const response = await fetch(`${this.baseUrl}/api/v1/event-types/${id}`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<Record<string, unknown>>(response);
    return transformEventType(data);
  }

  async list(): Promise<EventType[]> {
    const response = await fetch(`${this.baseUrl}/api/v1/event-types`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<{ event_types: Record<string, unknown>[] }>(response);
    return (data.event_types || []).map(transformEventType);
  }
}

class EventsResource {
  constructor(private baseUrl: string, private timeout: number) {}

  async create(request: CreateEventRequest): Promise<{ status: string; message: string }> {
    const response = await fetch(`${this.baseUrl}/api/v1/events`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal: AbortSignal.timeout(this.timeout),
    });
    return handleResponse<{ status: string; message: string }>(response);
  }

  async get(id: string): Promise<Event> {
    const response = await fetch(`${this.baseUrl}/api/v1/events/${id}`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<Record<string, unknown>>(response);
    return transformEvent(data);
  }

  async list(scopeId: string): Promise<Event[]> {
    const response = await fetch(`${this.baseUrl}/api/v1/events?scope_id=${scopeId}`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<{ events: Record<string, unknown>[] }>(response);
    return (data.events || []).map(transformEvent);
  }

  async getLineage(id: string): Promise<Lineage> {
    const response = await fetch(`${this.baseUrl}/api/v1/events/${id}/lineage`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<{
      event_id: string;
      parents: Record<string, unknown>[];
      children: Record<string, unknown>[];
    }>(response);
    return {
      event_id: data.event_id,
      parents: (data.parents || []).map(transformEvent),
      children: (data.children || []).map(transformEvent),
    };
  }
}

class ArtifactsResource {
  constructor(private baseUrl: string, private timeout: number) {}

  async create(request: CreateArtifactRequest): Promise<Artifact> {
    const response = await fetch(`${this.baseUrl}/api/v1/artifacts`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<Record<string, unknown>>(response);
    return transformArtifact(data);
  }

  async get(id: string): Promise<Artifact> {
    const response = await fetch(`${this.baseUrl}/api/v1/artifacts/${id}`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<Record<string, unknown>>(response);
    return transformArtifact(data);
  }

  async list(scopeId: string): Promise<Artifact[]> {
    const response = await fetch(`${this.baseUrl}/api/v1/artifacts?scope_id=${scopeId}`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<{ artifacts: Record<string, unknown>[] }>(response);
    return (data.artifacts || []).map(transformArtifact);
  }

  async linkToEvent(eventId: string, request: LinkArtifactRequest): Promise<{ status: string }> {
    const response = await fetch(`${this.baseUrl}/api/v1/events/${eventId}/artifacts`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal: AbortSignal.timeout(this.timeout),
    });
    return handleResponse<{ status: string }>(response);
  }

  async getForEvent(eventId: string): Promise<Artifact[]> {
    const response = await fetch(`${this.baseUrl}/api/v1/events/${eventId}/artifacts`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<{ artifacts: Record<string, unknown>[] }>(response);
    return (data.artifacts || []).map(transformArtifact);
  }
}

class ScoresResource {
  constructor(private baseUrl: string, private timeout: number) {}

  async create(eventId: string, request: CreateScoreRequest): Promise<Score> {
    const response = await fetch(`${this.baseUrl}/api/v1/events/${eventId}/scores`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<Record<string, unknown>>(response);
    return transformScore(data);
  }

  async list(eventId: string, type?: ScoreType): Promise<Score[]> {
    let url = `${this.baseUrl}/api/v1/events/${eventId}/scores`;
    if (type) url += `?type=${type}`;
    const response = await fetch(url, {
      signal: AbortSignal.timeout(this.timeout),
    });
    const data = await handleResponse<{ scores: Record<string, unknown>[] }>(response);
    return (data.scores || []).map(transformScore);
  }
}

/**
 * Low-level Lineage client with full API access.
 */
export class LineageClient {
  public readonly scopes: ScopesResource;
  public readonly actors: ActorsResource;
  public readonly eventTypes: EventTypesResource;
  public readonly events: EventsResource;
  public readonly artifacts: ArtifactsResource;
  public readonly scores: ScoresResource;

  private baseUrl: string;
  private timeout: number;

  constructor(config: LineageConfig = {}) {
    this.baseUrl = config.baseUrl || 'http://localhost:8080';
    this.timeout = config.timeout || 30000;

    this.scopes = new ScopesResource(this.baseUrl, this.timeout);
    this.actors = new ActorsResource(this.baseUrl, this.timeout);
    this.eventTypes = new EventTypesResource(this.baseUrl, this.timeout);
    this.events = new EventsResource(this.baseUrl, this.timeout);
    this.artifacts = new ArtifactsResource(this.baseUrl, this.timeout);
    this.scores = new ScoresResource(this.baseUrl, this.timeout);
  }

  async health(): Promise<HealthStatus> {
    const response = await fetch(`${this.baseUrl}/health`, {
      signal: AbortSignal.timeout(this.timeout),
    });
    return handleResponse<HealthStatus>(response);
  }
}

export { LineageError, NotFoundError, ValidationError, ServerError };
