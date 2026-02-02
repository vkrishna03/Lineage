"""
RAG Pipeline for Legal Documents with Lineage Tracking

This example demonstrates:
- Document ingestion with artifact tracking
- RAG retrieval with provenance
- AI-generated legal analysis (suggestion)
- Human lawyer review (decision)
- Full lineage chain from documents to final opinion

Run with: uv run python main.py
Requires: OPENAI_API_KEY env var (or uses mock responses)
"""

import os
import hashlib
from typing import Optional

import lineage

# Check for OpenAI API key
HAS_OPENAI = bool(os.getenv("OPENAI_API_KEY"))

if HAS_OPENAI:
    from langchain_openai import OpenAIEmbeddings, ChatOpenAI
    from langchain_community.vectorstores import Chroma
else:
    print("Note: OPENAI_API_KEY not set, using mock responses")


# Sample legal documents for the demo
LEGAL_DOCUMENTS = [
    {
        "id": "doc-001",
        "title": "Employment Contract Template",
        "content": """
        EMPLOYMENT AGREEMENT

        Non-Compete Clause: Employee agrees not to engage in any business activity
        that competes with the Company for a period of 12 months following termination.
        Geographic scope: Within 50 miles of any Company office.

        Severance: In case of termination without cause, Employee shall receive
        2 weeks of base salary for each year of employment, up to 26 weeks maximum.
        """,
        "type": "contract",
        "jurisdiction": "California",
    },
    {
        "id": "doc-002",
        "title": "California Non-Compete Law Summary",
        "content": """
        California Business and Professions Code Section 16600:

        Non-compete agreements are generally VOID and UNENFORCEABLE in California.
        Exceptions exist for sale of business goodwill and partnership dissolution.

        Recent case law (2023): Edwards v. Arthur Andersen LLP confirmed broad
        interpretation - even narrow non-competes are unenforceable.
        """,
        "type": "statute",
        "jurisdiction": "California",
    },
    {
        "id": "doc-003",
        "title": "Recent Court Ruling: TechCorp v. Smith (2024)",
        "content": """
        RULING SUMMARY:

        The court held that the non-compete clause in Smith's employment contract
        was unenforceable under California law, despite the clause being limited
        to 6 months and a 25-mile radius.

        Key finding: California's policy against non-competes applies regardless
        of the scope or duration of the restriction.

        Damages awarded: $150,000 for wrongful termination based on alleged
        violation of void non-compete.
        """,
        "type": "case_law",
        "jurisdiction": "California",
    },
]


def compute_hash(content: str) -> str:
    """Compute SHA-256 hash of document content."""
    return f"sha256:{hashlib.sha256(content.encode()).hexdigest()[:16]}"


class MockVectorStore:
    """Mock vector store for demo without OpenAI."""

    def __init__(self, documents: list[dict]):
        self.documents = documents

    def similarity_search(self, query: str, k: int = 3) -> list:
        """Return all documents as mock search results."""
        class MockDoc:
            def __init__(self, content: str, metadata: dict):
                self.page_content = content
                self.metadata = metadata

        return [
            MockDoc(doc["content"], {"id": doc["id"], "title": doc["title"]})
            for doc in self.documents[:k]
        ]


class LegalRAGPipeline:
    """RAG pipeline for legal document analysis with Lineage tracking."""

    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.vector_store: Optional[Chroma] = None
        self.ingested_docs: dict[str, str] = {}  # doc_id -> event_id

    async def initialize(self):
        """Initialize Lineage and vector store."""
        await lineage.init(
            project="legal-rag",
            domain="compliance",
            environment="demo",
            base_url=self.base_url,
            actor_name="Legal RAG System",
            actor_type="service",
        )

        if HAS_OPENAI:
            self.embeddings = OpenAIEmbeddings()
            self.llm = ChatOpenAI(model="gpt-4", temperature=0)
            self.vector_store = Chroma(
                embedding_function=self.embeddings,
                persist_directory="./chroma_db"
            )
        else:
            self.vector_store = MockVectorStore(LEGAL_DOCUMENTS)

    async def ingest_document(self, doc: dict) -> str:
        """Ingest a document and track with Lineage."""
        content_hash = compute_hash(doc["content"])

        # Track document ingestion as assertion (high confidence - it's a fact)
        event = lineage.emit(
            "document_ingestion",
            "assertion",
            {
                "document_id": doc["id"],
                "title": doc["title"],
                "type": doc["type"],
                "jurisdiction": doc["jurisdiction"],
                "content_hash": content_hash,
            },
            confidence=0.99,
            actor=("service", "Document Ingestion Service"),
        )

        # Add to vector store
        if HAS_OPENAI and isinstance(self.vector_store, Chroma):
            self.vector_store.add_texts(
                texts=[doc["content"]],
                metadatas=[{"id": doc["id"], "title": doc["title"]}],
                ids=[doc["id"]],
            )

        event_id = event.id if event else doc["id"]
        self.ingested_docs[doc["id"]] = event_id
        print(f"  Ingested: {doc['title']} (hash: {content_hash[:20]}...)")

        return event_id

    async def retrieve_relevant_docs(self, query: str, case_id: str) -> tuple[list[dict], str]:
        """Retrieve relevant documents and track retrieval."""
        # Perform similarity search
        results = self.vector_store.similarity_search(query, k=3)

        retrieved_docs = []
        source_event_ids = []

        for doc in results:
            doc_id = doc.metadata.get("id", "unknown")
            retrieved_docs.append({
                "id": doc_id,
                "title": doc.metadata.get("title", "Unknown"),
                "content_snippet": doc.page_content[:200] + "...",
            })
            if doc_id in self.ingested_docs:
                source_event_ids.append(self.ingested_docs[doc_id])

        # Track retrieval as exploration (gathering information)
        event = lineage.emit(
            "document_retrieval",
            "exploration",
            {
                "case_id": case_id,
                "query": query,
                "retrieved_count": len(retrieved_docs),
                "document_ids": [d["id"] for d in retrieved_docs],
            },
            confidence=0.85,  # Retrieval relevance is probabilistic
            actor=("tool", "Vector Search"),
        )

        event_id = event.id if event else "retrieval"
        print(f"  Retrieved {len(retrieved_docs)} relevant documents")

        return retrieved_docs, event_id

    async def generate_analysis(
        self,
        query: str,
        docs: list[dict],
        case_id: str,
        retrieval_event_id: str
    ) -> tuple[dict, str]:
        """Generate AI legal analysis."""

        # Build context from retrieved docs
        context = "\n\n".join([
            f"Document: {d['title']}\n{d['content_snippet']}"
            for d in docs
        ])

        if HAS_OPENAI:
            prompt = f"""You are a legal research assistant analyzing documents for a compliance review.

Based on the following documents:
{context}

Question: {query}

Provide a brief legal analysis with:
1. Key findings from the documents
2. Risk assessment (low/medium/high)
3. Recommended action

Be concise and cite specific documents."""

            response = self.llm.invoke(prompt)
            analysis_text = response.content
            risk_level = "high" if "unenforceable" in analysis_text.lower() else "medium"
        else:
            # Mock analysis for demo
            analysis_text = """Based on the retrieved documents:

1. KEY FINDINGS:
   - California Business Code Section 16600 renders non-compete agreements void
   - Recent case TechCorp v. Smith (2024) confirms even narrow non-competes are unenforceable
   - The employment contract's non-compete clause directly conflicts with California law

2. RISK ASSESSMENT: HIGH
   - Attempting to enforce the non-compete could result in liability
   - Precedent shows damages of $150,000 in similar cases

3. RECOMMENDED ACTION:
   - Do NOT enforce the non-compete clause
   - Consult employment counsel before any termination-related actions
   - Consider revising employment contract templates for California employees"""
            risk_level = "high"

        analysis = {
            "case_id": case_id,
            "query": query,
            "analysis": analysis_text,
            "risk_level": risk_level,
            "cited_documents": [d["id"] for d in docs],
            "confidence": 0.72,  # AI analysis confidence
        }

        # Track AI analysis as suggestion (proposing interpretation)
        event = lineage.emit(
            "legal_analysis",
            "suggestion",
            analysis,
            confidence=0.72,
            actor=("llm", "GPT-4 Legal Analyst"),
            parent=retrieval_event_id,
        )

        event_id = event.id if event else "analysis"
        print(f"  Generated analysis (risk: {risk_level})")

        return analysis, event_id

    async def lawyer_review(
        self,
        analysis: dict,
        analysis_event_id: str,
        approved: bool,
        lawyer_notes: str,
    ) -> tuple[dict, str]:
        """Human lawyer reviews and decides on AI analysis."""

        decision = {
            "case_id": analysis["case_id"],
            "ai_risk_assessment": analysis["risk_level"],
            "approved": approved,
            "lawyer_notes": lawyer_notes,
            "final_recommendation": "do_not_enforce" if approved else "requires_further_review",
        }

        # Track human decision (high confidence - human judgment)
        event = lineage.emit(
            "legal_review",
            "decision",
            decision,
            confidence=0.95,
            actor=("human", "Sarah Chen, Senior Counsel"),
            parent=analysis_event_id,
            reason="Reviewed AI analysis and supporting documents",
        )

        event_id = event.id if event else "review"
        status = "APPROVED" if approved else "REQUIRES REVIEW"
        print(f"  Lawyer review: {status}")

        return decision, event_id

    async def execute_recommendation(
        self,
        decision: dict,
        decision_event_id: str,
    ) -> str:
        """Execute the final recommendation."""

        execution = {
            "case_id": decision["case_id"],
            "action_taken": decision["final_recommendation"],
            "policy_updated": True,
            "notification_sent": True,
            "effective_date": "2024-01-20",
        }

        # Track execution (certain - action was taken)
        event = lineage.emit(
            "recommendation_execution",
            "execution",
            execution,
            confidence=1.0,
            actor=("service", "Compliance System"),
            parent=decision_event_id,
        )

        event_id = event.id if event else "execution"
        print(f"  Executed: {decision['final_recommendation']}")

        return event_id


async def main():
    """Run the RAG legal analysis demo."""
    print("\n" + "=" * 60)
    print("Legal RAG Pipeline with Lineage Tracking")
    print("=" * 60)

    pipeline = LegalRAGPipeline()
    await pipeline.initialize()

    # Phase 1: Ingest documents
    print("\n[Phase 1] Document Ingestion")
    print("-" * 40)
    for doc in LEGAL_DOCUMENTS:
        await pipeline.ingest_document(doc)

    # Phase 2: Process a legal query
    case_id = "CASE-2024-001"
    query = "Is the non-compete clause in our California employment contracts enforceable?"

    print(f"\n[Phase 2] Processing Query")
    print("-" * 40)
    print(f"Case ID: {case_id}")
    print(f"Query: {query}")

    # Retrieve relevant documents
    print("\n[Phase 3] Document Retrieval")
    print("-" * 40)
    docs, retrieval_event_id = await pipeline.retrieve_relevant_docs(query, case_id)

    # Generate AI analysis
    print("\n[Phase 4] AI Analysis")
    print("-" * 40)
    analysis, analysis_event_id = await pipeline.generate_analysis(
        query, docs, case_id, retrieval_event_id
    )

    # Lawyer review
    print("\n[Phase 5] Lawyer Review")
    print("-" * 40)
    decision, decision_event_id = await pipeline.lawyer_review(
        analysis,
        analysis_event_id,
        approved=True,
        lawyer_notes="AI analysis is accurate. California law is clear on this. Recommend immediate policy update.",
    )

    # Execute recommendation
    print("\n[Phase 6] Execute Recommendation")
    print("-" * 40)
    await pipeline.execute_recommendation(decision, decision_event_id)

    # Summary
    print("\n" + "=" * 60)
    print("Pipeline Complete - Lineage Tracked")
    print("=" * 60)
    print("""
Event Chain:
  [document_ingestion] x3 (assertion, confidence: 0.99)
           ↓
  [document_retrieval] (exploration, confidence: 0.85)
           ↓
  [legal_analysis] (suggestion, confidence: 0.72)
           ↓
  [legal_review] (decision, confidence: 0.95)
           ↓
  [recommendation_execution] (execution, confidence: 1.0)

View the full lineage at: http://localhost:8080/swagger/index.html
Query: GET /api/v1/events?scope_id=<scope_id>
""")


if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
