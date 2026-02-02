-- Enums for the Epistemic Transparency & Event Lineage System

CREATE TYPE actor_type AS ENUM (
    'human',
    'llm',
    'agent',
    'service',
    'tool'
);

CREATE TYPE event_intent AS ENUM (
    'exploration',  -- Low-commitment probing or brainstorming
    'suggestion',   -- A proposed course of action, not yet committed
    'assertion',    -- A claim about the state of the world
    'decision',     -- A committed choice by an accountable actor
    'execution'     -- A real-world side effect was triggered
);

CREATE TYPE score_type AS ENUM (
    'confidence',   -- How certain the actor is about this event
    'relevance',    -- How relevant to the target context
    'reliability',  -- Downstream-assessed reliability
    'agreement'     -- Degree of agreement with prior events
);

CREATE TYPE score_category AS ENUM (
    'very_low',
    'low',
    'moderate',
    'high',
    'very_high'
);

CREATE TYPE correction_type AS ENUM (
    'supersede',    -- Fully replaces the corrected event
    'amend',        -- Partially modifies or extends
    'retract'       -- Withdraws the corrected event; no replacement
);

CREATE TYPE artifact_role AS ENUM (
    'input',
    'output'
);
