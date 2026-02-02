/**
 * Global state management for simple API
 */

import { LineageClient } from './client';
import type {
  Actor,
  ActorType,
  Event,
  EventType,
  InitConfig,
  Intent,
  Scope,
} from './types';

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

class GlobalState {
  private initialized = false;
  private client: LineageClient | null = null;
  private scope: Scope | null = null;
  private defaultActor: Actor | null = null;
  private eventTypes: Map<string, EventType> = new Map();
  private actors: Map<string, Actor> = new Map();
  private lastEvent: Event | null = null;
  private waitTime = 2000;

  async init(config: InitConfig): Promise<void> {
    this.client = new LineageClient({ baseUrl: config.baseUrl });
    this.waitTime = (config.waitTime ?? 2) * 1000;

    if (config.autoCreate !== false) {
      // Create scope
      this.scope = await this.client.scopes.create({
        project: config.project,
        domain: config.domain,
        environment: config.environment,
      });

      // Create default actor if specified
      if (config.actorName) {
        this.defaultActor = await this.getOrCreateActor(
          config.actorType || 'service',
          config.actorName
        );
      }
    }

    this.initialized = true;
  }

  private ensureInitialized(): void {
    if (!this.initialized) {
      throw new Error('lineage.init() must be called first');
    }
  }

  private async getOrCreateActor(
    actorType: ActorType,
    actorName: string
  ): Promise<Actor> {
    const key = `${actorType}:${actorName}`;
    const existing = this.actors.get(key);
    if (existing) return existing;

    const actor = await this.client!.actors.create({
      type: actorType,
      name: actorName,
    });
    this.actors.set(key, actor);
    return actor;
  }

  private async getOrCreateEventType(
    name: string,
    allowedIntents?: Intent[]
  ): Promise<EventType> {
    const existing = this.eventTypes.get(name);
    if (existing) return existing;

    const eventType = await this.client!.eventTypes.create({
      name,
      version: '1.0',
      allowed_intents: allowedIntents,
    });
    this.eventTypes.set(name, eventType);
    return eventType;
  }

  private async resolveActor(
    actor?: [ActorType, string] | Actor
  ): Promise<Actor> {
    if (!actor) {
      if (!this.defaultActor) {
        throw new Error('No actor specified and no default actor set');
      }
      return this.defaultActor;
    }

    if (Array.isArray(actor)) {
      return this.getOrCreateActor(actor[0], actor[1]);
    }

    return actor;
  }

  async emit(
    eventType: string,
    intent: Intent,
    payload: Record<string, unknown>,
    options: {
      confidence?: number;
      actor?: [ActorType, string] | Actor;
      parent?: Event | string;
      reason?: string;
      wait?: boolean;
    } = {}
  ): Promise<Event | null> {
    this.ensureInitialized();

    // Resolve actor
    const resolvedActor = await this.resolveActor(options.actor);

    // Get or create event type
    const et = await this.getOrCreateEventType(eventType, [intent]);

    // Resolve parent
    let parentIds: string[] | undefined;
    if (options.parent) {
      parentIds = [
        typeof options.parent === 'string'
          ? options.parent
          : options.parent.id,
      ];
    }

    // Extract confidence from payload if present
    let confidence = options.confidence;
    const payloadCopy = { ...payload };
    if (confidence === undefined && 'confidence' in payloadCopy) {
      confidence = payloadCopy.confidence as number;
      delete payloadCopy.confidence;
    }

    // Create event
    await this.client!.events.create({
      scope_id: this.scope!.id,
      actor_id: resolvedActor.id,
      event_type_id: et.id,
      intent,
      payload: payloadCopy,
      confidence,
      parent_event_ids: parentIds,
      reason: options.reason,
    });

    // Wait for async processing and fetch event
    const shouldWait = options.wait !== false;
    if (shouldWait && this.waitTime > 0) {
      await sleep(this.waitTime);
      const events = await this.client!.events.list(this.scope!.id);
      for (const event of events) {
        if (
          event.intent === intent &&
          event.actor_id === resolvedActor.id
        ) {
          this.lastEvent = event;
          return event;
        }
      }
    }

    return null;
  }

  getClient(): LineageClient {
    this.ensureInitialized();
    return this.client!;
  }

  getScope(): Scope {
    this.ensureInitialized();
    return this.scope!;
  }

  getLastEvent(): Event | null {
    return this.lastEvent;
  }
}

export const globalState = new GlobalState();
