"""
AI-Assisted Pricing Decision - Simple API

This example demonstrates the Lineage SDK's simple API.
Compare this to the verbose low-level API - it's much cleaner!

Run with: uv run python main.py
"""

import uuid

import lineage

# Unique project name for this run
RUN_ID = str(uuid.uuid4())[:8]


def main():
    print("=" * 60)
    print("AI-Assisted Pricing Decision Workflow")
    print("=" * 60)

    # Initialize Lineage - one line!
    lineage.init(
        project=f"ecommerce-{RUN_ID}",
        domain="pricing",
        environment="production",
        wait_time=2.0,
    )

    print(f"\nProject: ecommerce-{RUN_ID}")

    # =========================================
    # STEP 1: Service ingests sales data
    # =========================================
    print("\n" + "=" * 60)
    print("STEP 1: Service ingests sales data")
    print("=" * 60)

    e1 = lineage.emit(
        "data_ingestion",
        "assertion",
        {"product_id": "SKU-12345", "current_price": 29.99, "units_sold": 1523},
        confidence=0.99,
        actor=("service", "Sales Data Pipeline"),
    )
    print("Data ingestion event created")
    print("  Intent: assertion")
    print("  Confidence: 0.99")

    # =========================================
    # STEP 2: AI analyzes and recommends
    # =========================================
    print("\n" + "=" * 60)
    print("STEP 2: AI model analyzes and recommends pricing")
    print("=" * 60)

    # Using decorator style
    @lineage.track("recommendation", intent="suggestion", actor=("llm", "Pricing Optimizer GPT"))
    def recommend_price(data: dict) -> dict:
        """AI recommends a price based on data analysis."""
        return {
            "product_id": data["product_id"],
            "recommended_price": 26.99,
            "reasoning": "Competitor analysis suggests price reduction",
            "confidence": 0.72,  # Extracted automatically!
        }

    rec = recommend_price({"product_id": "SKU-12345"})
    e2 = lineage.get_last_event()

    print("AI recommendation created")
    print("  Intent: suggestion")
    print(f"  Recommended: ${rec['recommended_price']}")
    print("  Confidence: 0.72")

    # =========================================
    # STEP 3: Human reviews and decides
    # =========================================
    print("\n" + "=" * 60)
    print("STEP 3: Human reviews and makes decision")
    print("=" * 60)

    # Using span style
    with lineage.span("recommendation", "decision", actor=("human", "Sarah Chen")) as span:
        # Human reviews and adjusts
        span.payload = {
            "product_id": "SKU-12345",
            "approved_price": 27.99,  # Adjusted from AI's $26.99
            "reasoning": "Adjusted to maintain minimum margin",
        }
        span.confidence = 0.88

    e3 = span.event
    print("Human decision created")
    print("  Intent: decision")
    print("  Approved: $27.99")
    print("  Confidence: 0.88")

    # =========================================
    # STEP 4: Service executes pricing change
    # =========================================
    print("\n" + "=" * 60)
    print("STEP 4: Service executes approved pricing change")
    print("=" * 60)

    e4 = lineage.emit(
        "execution",
        "execution",
        {
            "product_id": "SKU-12345",
            "old_price": 29.99,
            "new_price": 27.99,
            "effective_at": "2026-02-03T00:00:00Z",
        },
        confidence=1.0,
        actor=("service", "Pricing Engine"),
        parent=e3,  # Link to decision
    )
    print("Execution event created")
    print("  Intent: execution")
    print("  Price changed: $29.99 -> $27.99")
    print("  Confidence: 1.0")

    # =========================================
    # QUERY: Show lineage
    # =========================================
    print("\n" + "=" * 60)
    print("LINEAGE: Full decision trail")
    print("=" * 60)

    if e4:
        client = lineage.get_client()
        lin = client.events.get_lineage(e4.id)
        print(f"\nExecution event has {len(lin.parents)} parent(s)")

    print("\n" + "=" * 60)
    print("Workflow complete!")
    print("=" * 60)


if __name__ == "__main__":
    main()
