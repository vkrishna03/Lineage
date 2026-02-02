"""Enums for the Lineage SDK."""

from enum import Enum


class ActorType(str, Enum):
    """Types of actors that can generate events."""

    HUMAN = "human"
    LLM = "llm"
    AGENT = "agent"
    SERVICE = "service"
    TOOL = "tool"


class Intent(str, Enum):
    """Event intent - the epistemic status of an action."""

    EXPLORATION = "exploration"
    SUGGESTION = "suggestion"
    ASSERTION = "assertion"
    DECISION = "decision"
    EXECUTION = "execution"


class CorrectionType(str, Enum):
    """Types of corrections that can be made to events."""

    SUPERSEDE = "supersede"
    AMEND = "amend"
    RETRACT = "retract"


class ScoreType(str, Enum):
    """Types of scores that can be attached to events."""

    CONFIDENCE = "confidence"
    RELEVANCE = "relevance"
    RELIABILITY = "reliability"
    AGREEMENT = "agreement"


class ScoreCategory(str, Enum):
    """Categories derived from score values."""

    VERY_LOW = "very_low"
    LOW = "low"
    MODERATE = "moderate"
    HIGH = "high"
    VERY_HIGH = "very_high"


class ArtifactRole(str, Enum):
    """Role of an artifact in relation to an event."""

    INPUT = "input"
    OUTPUT = "output"
