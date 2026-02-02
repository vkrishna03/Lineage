"""
FastAPI Service with Lineage Tracking

This example demonstrates:
- Request-scoped Lineage tracking with middleware
- Multiple actors within a single request
- AI-assisted pricing with human override capability
- Error tracking and audit trails
- Production-ready patterns

Run with: uv run uvicorn main:app --reload
API docs at: http://localhost:8000/docs
"""

import random
from contextlib import asynccontextmanager
from typing import Optional
from uuid import uuid4

from fastapi import FastAPI, HTTPException, Request, Depends
from pydantic import BaseModel

import lineage


# ============================================================================
# Models
# ============================================================================

class PriceRequest(BaseModel):
    product_id: str
    current_price: float
    competitor_prices: list[float] = []
    inventory_level: int = 100


class PriceResponse(BaseModel):
    product_id: str
    recommended_price: float
    confidence: float
    reasoning: str
    request_id: str


class PriceApproval(BaseModel):
    request_id: str
    approved: bool
    override_price: Optional[float] = None
    reviewer_notes: str = ""


class PriceResult(BaseModel):
    product_id: str
    final_price: float
    approved_by: str
    execution_id: str


# ============================================================================
# In-memory storage (use Redis/DB in production)
# ============================================================================

pending_approvals: dict[str, dict] = {}


# ============================================================================
# Lifespan: Initialize Lineage on startup
# ============================================================================

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize Lineage SDK on startup."""
    await lineage.init(
        project="pricing-api",
        domain="pricing",
        environment="demo",
        base_url="http://localhost:8080",
        actor_name="Pricing API",
        actor_type="service",
        wait_time=1,  # Shorter wait for demo
    )
    print("Lineage SDK initialized")
    yield


app = FastAPI(
    title="Pricing API with Lineage",
    description="AI-assisted pricing service with full decision lineage",
    version="1.0.0",
    lifespan=lifespan,
)


# ============================================================================
# Request ID middleware
# ============================================================================

@app.middleware("http")
async def add_request_id(request: Request, call_next):
    """Add request ID for tracing."""
    request.state.request_id = str(uuid4())[:8]
    response = await call_next(request)
    response.headers["X-Request-ID"] = request.state.request_id
    return response


def get_request_id(request: Request) -> str:
    """Dependency to get request ID."""
    return request.state.request_id


# ============================================================================
# Mock AI pricing model
# ============================================================================

def mock_ai_pricing(
    current_price: float,
    competitor_prices: list[float],
    inventory_level: int,
) -> tuple[float, float, str]:
    """
    Mock AI pricing model.
    In production, this would call your ML model or LLM.
    """
    if competitor_prices:
        avg_competitor = sum(competitor_prices) / len(competitor_prices)
        price_position = current_price / avg_competitor

        if price_position > 1.1:  # 10% above competitors
            recommended = current_price * 0.95
            reasoning = f"Price is {price_position:.0%} of competitor average. Recommend 5% reduction."
            confidence = 0.75
        elif price_position < 0.9:  # 10% below competitors
            recommended = current_price * 1.05
            reasoning = f"Price is below market. Room for 5% increase."
            confidence = 0.68
        else:
            recommended = current_price
            reasoning = "Price is competitive. No change recommended."
            confidence = 0.82
    else:
        # No competitor data, use inventory heuristic
        if inventory_level > 150:
            recommended = current_price * 0.9
            reasoning = "High inventory. Recommend 10% discount to move stock."
            confidence = 0.6
        elif inventory_level < 50:
            recommended = current_price * 1.1
            reasoning = "Low inventory. Can support 10% premium."
            confidence = 0.65
        else:
            recommended = current_price
            reasoning = "Inventory levels normal. Maintain current price."
            confidence = 0.7

    return round(recommended, 2), confidence, reasoning


# ============================================================================
# Endpoints
# ============================================================================

@app.post("/api/v1/price/recommend", response_model=PriceResponse)
async def recommend_price(
    req: PriceRequest,
    request_id: str = Depends(get_request_id),
):
    """
    Get AI-recommended price for a product.

    This creates two Lineage events:
    1. Data ingestion (assertion) - recording the input data
    2. Price recommendation (suggestion) - AI's recommendation
    """
    # Track data ingestion (high confidence - these are facts)
    data_event = lineage.emit(
        "price_data_ingestion",
        "assertion",
        {
            "request_id": request_id,
            "product_id": req.product_id,
            "current_price": req.current_price,
            "competitor_prices": req.competitor_prices,
            "inventory_level": req.inventory_level,
        },
        confidence=0.99,
        actor=("service", "Pricing API"),
    )

    # Get AI recommendation
    recommended_price, confidence, reasoning = mock_ai_pricing(
        req.current_price,
        req.competitor_prices,
        req.inventory_level,
    )

    # Track AI suggestion (lower confidence - it's a prediction)
    parent_id = data_event.id if data_event else None
    suggestion_event = lineage.emit(
        "price_recommendation",
        "suggestion",
        {
            "request_id": request_id,
            "product_id": req.product_id,
            "current_price": req.current_price,
            "recommended_price": recommended_price,
            "reasoning": reasoning,
        },
        confidence=confidence,
        actor=("llm", "Pricing AI Model"),
        parent=parent_id,
    )

    # Store for approval workflow
    pending_approvals[request_id] = {
        "product_id": req.product_id,
        "current_price": req.current_price,
        "recommended_price": recommended_price,
        "confidence": confidence,
        "reasoning": reasoning,
        "suggestion_event_id": suggestion_event.id if suggestion_event else None,
    }

    return PriceResponse(
        product_id=req.product_id,
        recommended_price=recommended_price,
        confidence=confidence,
        reasoning=reasoning,
        request_id=request_id,
    )


@app.post("/api/v1/price/approve", response_model=PriceResult)
async def approve_price(approval: PriceApproval):
    """
    Human approval/rejection of price recommendation.

    This creates two Lineage events:
    1. Human decision - approval or rejection with optional override
    2. Execution - the actual price change (if approved)
    """
    if approval.request_id not in pending_approvals:
        raise HTTPException(status_code=404, detail="Request not found")

    pending = pending_approvals.pop(approval.request_id)

    # Determine final price
    if approval.approved:
        final_price = approval.override_price or pending["recommended_price"]
        was_overridden = approval.override_price is not None
    else:
        final_price = pending["current_price"]  # Keep current price
        was_overridden = False

    # Track human decision
    decision_event = lineage.emit(
        "price_decision",
        "decision",
        {
            "request_id": approval.request_id,
            "product_id": pending["product_id"],
            "approved": approval.approved,
            "ai_recommended_price": pending["recommended_price"],
            "final_price": final_price,
            "was_overridden": was_overridden,
            "reviewer_notes": approval.reviewer_notes,
        },
        confidence=0.95,  # Human decision is high confidence
        actor=("human", "Pricing Manager"),
        parent=pending.get("suggestion_event_id"),
        reason=approval.reviewer_notes or ("Approved AI recommendation" if approval.approved else "Rejected recommendation"),
    )

    if not approval.approved:
        return PriceResult(
            product_id=pending["product_id"],
            final_price=final_price,
            approved_by="Pricing Manager (REJECTED)",
            execution_id="none",
        )

    # Track execution (certain - price was changed)
    execution_event = lineage.emit(
        "price_execution",
        "execution",
        {
            "request_id": approval.request_id,
            "product_id": pending["product_id"],
            "old_price": pending["current_price"],
            "new_price": final_price,
            "change_percentage": round((final_price - pending["current_price"]) / pending["current_price"] * 100, 2),
        },
        confidence=1.0,  # Execution is certain
        actor=("service", "Pricing API"),
        parent=decision_event.id if decision_event else None,
    )

    execution_id = execution_event.id if execution_event else "mock-execution"

    return PriceResult(
        product_id=pending["product_id"],
        final_price=final_price,
        approved_by="Pricing Manager",
        execution_id=execution_id,
    )


@app.get("/api/v1/price/pending")
async def list_pending():
    """List all pending price approvals."""
    return {
        "pending": [
            {
                "request_id": k,
                "product_id": v["product_id"],
                "current_price": v["current_price"],
                "recommended_price": v["recommended_price"],
                "confidence": v["confidence"],
            }
            for k, v in pending_approvals.items()
        ]
    }


@app.get("/health")
async def health():
    """Health check."""
    return {"status": "healthy", "lineage": "connected"}


# ============================================================================
# Demo script
# ============================================================================

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
