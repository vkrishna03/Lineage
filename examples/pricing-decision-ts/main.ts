/**
 * Pricing Decision Workflow with Lineage Tracking (TypeScript)
 *
 * This example demonstrates:
 * - Multi-actor workflow (service → LLM → human → service)
 * - Intent progression (assertion → suggestion → decision → execution)
 * - Confidence scores for each step
 * - Parent-child event linking
 *
 * Run with: npm start (or npx tsx main.ts)
 * Requires: Lineage API server at http://localhost:8080
 */

import * as lineage from 'lineage-sdk';

// Types for our pricing workflow
interface ProductData {
  productId: string;
  currentPrice: number;
  competitorPrices: number[];
  unitsSold: number;
}

interface PriceRecommendation {
  productId: string;
  recommendedPrice: number;
  reasoning: string;
  confidence: number;
}

interface PriceDecision {
  productId: string;
  approvedPrice: number;
  originalRecommendation: number;
  wasModified: boolean;
  reviewerNotes: string;
}

// Mock AI pricing model
function aiPricingModel(data: ProductData): PriceRecommendation {
  const avgCompetitor =
    data.competitorPrices.reduce((a, b) => a + b, 0) / data.competitorPrices.length;

  let recommendedPrice: number;
  let reasoning: string;
  let confidence: number;

  if (data.currentPrice > avgCompetitor * 1.1) {
    recommendedPrice = Math.round(data.currentPrice * 0.95 * 100) / 100;
    reasoning = 'Price is 10%+ above competitor average. Recommend 5% reduction.';
    confidence = 0.75;
  } else if (data.currentPrice < avgCompetitor * 0.9) {
    recommendedPrice = Math.round(data.currentPrice * 1.05 * 100) / 100;
    reasoning = 'Price is below market. Room for 5% increase.';
    confidence = 0.68;
  } else {
    recommendedPrice = data.currentPrice;
    reasoning = 'Price is competitive. No change recommended.';
    confidence = 0.82;
  }

  return { productId: data.productId, recommendedPrice, reasoning, confidence };
}

async function main() {
  console.log('\n' + '='.repeat(60));
  console.log('Pricing Decision Workflow with Lineage (TypeScript)');
  console.log('='.repeat(60));

  // Initialize Lineage
  console.log('\n[Step 0] Initializing Lineage...');
  await lineage.init({
    project: 'pricing-demo-ts',
    domain: 'pricing',
    environment: 'demo',
    baseUrl: 'http://localhost:8080',
    actorName: 'Pricing Service',
    actorType: 'service',
    waitTime: 1, // Shorter wait for demo
  });
  console.log('  Lineage initialized');

  // Product data
  const productData: ProductData = {
    productId: 'SKU-123',
    currentPrice: 29.99,
    competitorPrices: [27.99, 28.50, 31.0, 26.99],
    unitsSold: 1500,
  };

  // Step 1: Service asserts data (high confidence fact)
  console.log('\n[Step 1] Data Ingestion (assertion)');
  console.log('-'.repeat(40));
  const dataEvent = await lineage.emit('price_data_ingestion', 'assertion', {
    productId: productData.productId,
    currentPrice: productData.currentPrice,
    competitorPrices: productData.competitorPrices,
    unitsSold: productData.unitsSold,
  }, {
    confidence: 0.99,
    actor: ['service', 'Data Pipeline'],
  });
  console.log(`  Ingested data for ${productData.productId}`);
  console.log(`  Current price: $${productData.currentPrice}`);

  // Step 2: AI suggests price (moderate confidence)
  console.log('\n[Step 2] AI Recommendation (suggestion)');
  console.log('-'.repeat(40));
  const recommendation = aiPricingModel(productData);

  const suggestionEvent = await lineage.emit('price_recommendation', 'suggestion', {
    productId: recommendation.productId,
    currentPrice: productData.currentPrice,
    recommendedPrice: recommendation.recommendedPrice,
    reasoning: recommendation.reasoning,
  }, {
    confidence: recommendation.confidence,
    actor: ['llm', 'Pricing AI'],
    parent: dataEvent ?? undefined,
  });
  console.log(`  Recommended: $${recommendation.recommendedPrice}`);
  console.log(`  Reasoning: ${recommendation.reasoning}`);
  console.log(`  Confidence: ${(recommendation.confidence * 100).toFixed(0)}%`);

  // Step 3: Human decides (high confidence)
  console.log('\n[Step 3] Human Decision (decision)');
  console.log('-'.repeat(40));

  // Simulate human adjusting the price slightly
  const humanDecision: PriceDecision = {
    productId: productData.productId,
    approvedPrice: 27.99, // Human chose a round number
    originalRecommendation: recommendation.recommendedPrice,
    wasModified: true,
    reviewerNotes: 'Adjusted to psychological price point ($27.99)',
  };

  const decisionEvent = await lineage.emit('price_decision', 'decision', {
    productId: humanDecision.productId,
    approvedPrice: humanDecision.approvedPrice,
    originalRecommendation: humanDecision.originalRecommendation,
    wasModified: humanDecision.wasModified,
    reviewerNotes: humanDecision.reviewerNotes,
  }, {
    confidence: 0.92,
    actor: ['human', 'Sarah Chen, Pricing Manager'],
    parent: suggestionEvent ?? undefined,
    reason: humanDecision.reviewerNotes,
  });
  console.log(`  Approved: $${humanDecision.approvedPrice}`);
  console.log(`  Modified: ${humanDecision.wasModified ? 'Yes' : 'No'}`);
  console.log(`  Notes: ${humanDecision.reviewerNotes}`);

  // Step 4: Service executes (certain)
  console.log('\n[Step 4] Execution (execution)');
  console.log('-'.repeat(40));

  await lineage.emit('price_execution', 'execution', {
    productId: productData.productId,
    oldPrice: productData.currentPrice,
    newPrice: humanDecision.approvedPrice,
    changePercent: Math.round((humanDecision.approvedPrice - productData.currentPrice) / productData.currentPrice * 100 * 100) / 100,
    effectiveAt: new Date().toISOString(),
  }, {
    confidence: 1.0,
    actor: ['service', 'Pricing Engine'],
    parent: decisionEvent ?? undefined,
  });
  console.log(`  Price changed: $${productData.currentPrice} → $${humanDecision.approvedPrice}`);
  console.log(`  Change: ${((humanDecision.approvedPrice - productData.currentPrice) / productData.currentPrice * 100).toFixed(1)}%`);

  // Summary
  console.log('\n' + '='.repeat(60));
  console.log('Workflow Complete - Lineage Tracked');
  console.log('='.repeat(60));
  console.log(`
Event Chain:
  [price_data_ingestion] (assertion, 0.99) - Data Pipeline
           ↓
  [price_recommendation] (suggestion, ${recommendation.confidence}) - Pricing AI
           ↓
  [price_decision] (decision, 0.92) - Sarah Chen
           ↓
  [price_execution] (execution, 1.0) - Pricing Engine

Actors involved:
  • Data Pipeline (service) - Ingested product data
  • Pricing AI (llm) - Recommended price change
  • Sarah Chen (human) - Approved with modification
  • Pricing Engine (service) - Applied the change

View at: http://localhost:8080/swagger/index.html
`);
}

// Run
main().catch(console.error);
