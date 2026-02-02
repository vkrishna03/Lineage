/**
 * Lineage SDK - Epistemic transparency for AI systems
 *
 * Simple API:
 *   import * as lineage from 'lineage-sdk';
 *
 *   await lineage.init({ project: 'my-app', actorName: 'my-service' });
 *   await lineage.emit('data_ingestion', 'assertion', { data: 'loaded' }, { confidence: 0.99 });
 *
 * Low-level API:
 *   import { LineageClient } from 'lineage-sdk';
 *   const client = new LineageClient({ baseUrl: 'http://localhost:8080' });
 */

import { globalState } from './state';
import type {
  ActorType,
  Event,
  InitConfig,
  Intent,
  Scope,
  TrackOptions,
} from './types';

// Re-export client and types
export { LineageClient, LineageError, NotFoundError, ValidationError, ServerError } from './client';
export type * from './types';

/**
 * Initialize the Lineage SDK.
 *
 * @example
 * await lineage.init({
 *   project: 'my-app',
 *   actorName: 'my-service',
 *   actorType: 'service'
 * });
 */
export async function init(config: InitConfig): Promise<void> {
  await globalState.init(config);
}

/**
 * Emit a Lineage event.
 *
 * @example
 * await lineage.emit('data_ingestion', 'assertion', { data: 'loaded' }, { confidence: 0.99 });
 */
export async function emit(
  eventType: string,
  intent: Intent,
  payload: Record<string, unknown>,
  options: {
    confidence?: number;
    actor?: [ActorType, string];
    parent?: Event | string;
    reason?: string;
    wait?: boolean;
  } = {}
): Promise<Event | null> {
  return globalState.emit(eventType, intent, payload, options);
}

/**
 * Create a tracked wrapper function.
 *
 * @example
 * const recommend = lineage.track('recommendation', 'suggestion', { actor: ['llm', 'GPT-4'] })(
 *   async (data) => {
 *     return { price: 26.99, confidence: 0.85 };
 *   }
 * );
 */
export function track<T extends Record<string, unknown>, R extends Record<string, unknown>>(
  eventType: string,
  intent: Intent,
  options: TrackOptions = {}
) {
  return function <F extends (input: T) => R | Promise<R>>(fn: F): F {
    const wrapped = async function (input: T): Promise<R> {
      const result = await fn(input);

      // Extract confidence from result if present
      let confidence = options.confidence;
      const payload = { ...result };
      if (confidence === undefined && 'confidence' in payload) {
        confidence = payload.confidence as number;
        delete payload.confidence;
      }

      await globalState.emit(eventType, intent, payload, {
        confidence,
        actor: options.actor,
      });

      return result;
    };

    return wrapped as F;
  };
}

/**
 * Get the underlying LineageClient for advanced operations.
 */
export function getClient() {
  return globalState.getClient();
}

/**
 * Get the current scope.
 */
export function getScope(): Scope {
  return globalState.getScope();
}

/**
 * Get the most recently emitted event.
 */
export function getLastEvent(): Event | null {
  return globalState.getLastEvent();
}
