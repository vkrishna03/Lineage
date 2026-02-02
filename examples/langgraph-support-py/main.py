"""
LangGraph Customer Support Agent with Lineage Tracking

This example demonstrates a production-style customer support agent that:
1. Analyzes customer queries (exploration)
2. Generates AI responses (suggestion)
3. Escalates complex issues to humans (human decision)
4. Sends final response (execution)

All decisions are tracked with full lineage for audit and debugging.

Requirements:
- OPENAI_API_KEY environment variable
- Lineage API server running at localhost:8080

Run with: uv run python main.py
"""

import os
import uuid
from typing import Annotated, TypedDict

import lineage
from langchain_core.messages import HumanMessage, AIMessage
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages

# Check for OpenAI API key
if not os.environ.get("OPENAI_API_KEY"):
    print("Warning: OPENAI_API_KEY not set. Using mock responses.")
    USE_MOCK = True
else:
    USE_MOCK = False

# Unique run ID
RUN_ID = str(uuid.uuid4())[:8]


# State definition
class SupportState(TypedDict):
    """State for the customer support workflow."""
    messages: Annotated[list, add_messages]
    ticket_id: str
    query: str
    sentiment: str
    needs_escalation: bool
    response: str
    lineage_events: list  # Track events for lineage


def mock_llm_response(prompt: str, task: str) -> str:
    """Mock LLM responses for testing without API key."""
    if task == "sentiment":
        if "angry" in prompt.lower() or "furious" in prompt.lower() or "terrible" in prompt.lower():
            return "negative"
        elif "broken" in prompt.lower() or "not working" in prompt.lower():
            return "negative"
        return "positive"
    elif task == "response":
        return "Thank you for contacting us. I understand your concern and I'm here to help. Let me look into this for you."
    return "OK"


# Node functions
def analyze_sentiment(state: SupportState) -> SupportState:
    """
    Analyze customer sentiment to determine escalation needs.
    Intent: exploration (gathering information, no commitment)
    """
    query = state["query"]

    if USE_MOCK:
        sentiment = mock_llm_response(query, "sentiment")
    else:
        llm = ChatOpenAI(model="gpt-4o-mini", temperature=0)
        result = llm.invoke(f"""
            Analyze the sentiment of this customer message.
            Reply with exactly one word: "positive", "neutral", or "negative"

            Message: {query}
        """)
        sentiment = result.content.strip().lower()

    # Track with Lineage - exploration intent
    event = lineage.emit(
        "sentiment_analysis",
        "exploration",
        {
            "ticket_id": state["ticket_id"],
            "query_preview": query[:100],
            "sentiment": sentiment,
            "analysis_model": "gpt-4o-mini" if not USE_MOCK else "mock",
        },
        confidence=0.85,
        actor=("agent", "Support Agent"),
    )

    needs_escalation = sentiment == "negative"

    return {
        **state,
        "sentiment": sentiment,
        "needs_escalation": needs_escalation,
        "lineage_events": [event],
    }


def generate_response(state: SupportState) -> SupportState:
    """
    Generate an AI response to the customer query.
    Intent: suggestion (AI is proposing, not deciding)
    """
    query = state["query"]
    sentiment = state["sentiment"]

    if USE_MOCK:
        response = mock_llm_response(query, "response")
    else:
        llm = ChatOpenAI(model="gpt-4o-mini", temperature=0.7)
        result = llm.invoke(f"""
            You are a helpful customer support agent.
            The customer seems {sentiment}. Respond appropriately.

            Customer: {query}

            Provide a helpful, empathetic response.
        """)
        response = result.content

    # Track with Lineage - suggestion intent (AI proposing, not final)
    event = lineage.emit(
        "response_draft",
        "suggestion",
        {
            "ticket_id": state["ticket_id"],
            "draft_response": response[:500],
            "sentiment_context": sentiment,
            "auto_approved": not state["needs_escalation"],
        },
        confidence=0.72,  # Lower confidence - it's a suggestion
        actor=("llm", "GPT-4o-mini"),
        parent=state["lineage_events"][-1],
    )

    return {
        **state,
        "response": response,
        "lineage_events": state["lineage_events"] + [event],
    }


def human_review(state: SupportState) -> SupportState:
    """
    Human reviews escalated tickets and may modify the response.
    Intent: decision (human making the final call)
    """
    original_response = state["response"]

    # Simulate human review (in production: webhook, queue, UI)
    # Human decides to personalize the response more
    human_response = f"""I sincerely apologize for the frustration you're experiencing.
I've reviewed your case personally and I want to make this right.

{original_response}

I'm also adding a 20% discount to your next order as a gesture of goodwill.
Please let me know if there's anything else I can do.

Best regards,
Sarah - Customer Support Lead"""

    # Track with Lineage - decision intent (human authority)
    event = lineage.emit(
        "human_review",
        "decision",
        {
            "ticket_id": state["ticket_id"],
            "original_draft": original_response[:200],
            "approved_response": human_response[:200],
            "human_modified": True,
            "compensation_offered": "20% discount",
        },
        confidence=0.95,  # High confidence - human decision
        actor=("human", "Sarah Chen, Support Lead"),
        parent=state["lineage_events"][-1],
    )

    return {
        **state,
        "response": human_response,
        "lineage_events": state["lineage_events"] + [event],
    }


def send_response(state: SupportState) -> SupportState:
    """
    Send the final response to the customer.
    Intent: execution (action taken)
    """
    # Track with Lineage - execution intent
    lineage.emit(
        "response_sent",
        "execution",
        {
            "ticket_id": state["ticket_id"],
            "channel": "email",
            "response_length": len(state["response"]),
            "was_escalated": state["needs_escalation"],
        },
        confidence=1.0,  # Execution is deterministic
        actor=("service", "Email Service"),
        parent=state["lineage_events"][-1],
    )

    return state


def should_escalate(state: SupportState) -> str:
    """Route to human review if escalation needed."""
    return "human_review" if state["needs_escalation"] else "send_response"


def build_graph() -> StateGraph:
    """Build the LangGraph workflow."""
    graph = StateGraph(SupportState)

    # Add nodes
    graph.add_node("analyze", analyze_sentiment)
    graph.add_node("generate", generate_response)
    graph.add_node("human_review", human_review)
    graph.add_node("send", send_response)

    # Define edges
    graph.set_entry_point("analyze")
    graph.add_edge("analyze", "generate")
    graph.add_conditional_edges("generate", should_escalate)
    graph.add_edge("human_review", "send")
    graph.add_edge("send", END)

    return graph.compile()


def process_ticket(ticket_id: str, query: str):
    """Process a customer support ticket."""
    print(f"\n{'='*60}")
    print(f"Processing Ticket: {ticket_id}")
    print(f"{'='*60}")
    print(f"Query: {query[:100]}...")

    initial_state = SupportState(
        messages=[HumanMessage(content=query)],
        ticket_id=ticket_id,
        query=query,
        sentiment="",
        needs_escalation=False,
        response="",
        lineage_events=[],
    )

    app = build_graph()
    final_state = app.invoke(initial_state)

    print(f"\nSentiment: {final_state['sentiment']}")
    print(f"Escalated: {final_state['needs_escalation']}")
    print(f"\nFinal Response:")
    print("-" * 40)
    print(final_state["response"][:300])
    if len(final_state["response"]) > 300:
        print("...")
    print("-" * 40)

    return final_state


def main():
    print("=" * 60)
    print("LangGraph Customer Support with Lineage Tracking")
    print("=" * 60)

    # Initialize Lineage
    lineage.init(
        project=f"customer-support-{RUN_ID}",
        domain="support-tickets",
        environment="demo",
        wait_time=2.0,
    )
    print(f"\nLineage project: customer-support-{RUN_ID}")

    # Test Case 1: Simple query (no escalation)
    print("\n" + "=" * 60)
    print("TEST CASE 1: Simple Query (Auto-Response)")
    print("=" * 60)

    process_ticket(
        "TICKET-001",
        "Hi, I'd like to know what time your store closes on weekends."
    )

    # Test Case 2: Angry customer (escalation)
    print("\n" + "=" * 60)
    print("TEST CASE 2: Angry Customer (Human Escalation)")
    print("=" * 60)

    process_ticket(
        "TICKET-002",
        "I am FURIOUS! My order arrived broken and nobody is helping me. "
        "This is the WORST customer service I've ever experienced. "
        "I want a refund immediately!"
    )

    # Show lineage summary
    print("\n" + "=" * 60)
    print("LINEAGE SUMMARY")
    print("=" * 60)

    client = lineage.get_client()
    scope = lineage.get_scope()
    events = client.events.list(scope_id=scope.id)

    print(f"\nTotal events tracked: {len(events)}")
    print("\nEvent chain:")
    for e in events:
        actor = client.actors.get(e.actor_id)
        scores = client.scores.list(event_id=e.id)
        conf = next((s.value for s in scores if s.type.value == "confidence"), None)
        conf_str = f" (conf: {conf:.2f})" if conf else ""
        print(f"  [{actor.type.value}] {e.intent.value}: {actor.name}{conf_str}")

    print("\n" + "=" * 60)
    print("Demo complete! Check Lineage API for full audit trail.")
    print("=" * 60)


if __name__ == "__main__":
    main()
