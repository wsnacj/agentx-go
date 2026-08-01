---
name: a-stock-data
description: Use for A-share market-data, valuation snapshot, research-report, announcement, signal, hot-topic, Dragon Tiger Board, lockup-expiry, industry-rotation, company-profile, and bounded multi-stock comparison requests. Prefer structured AgentX tools over free-form endpoint snippets.
---

# A-Stock Data Skill

Use this skill for A-share tasks such as quote lookup, valuation snapshots, research report discovery, announcements, company profile, hot-topic attribution, capital-flow checks, Dragon Tiger Board tracking, lockup expiry, industry comparison, and bounded batch investigation.

## Tool Preference

Prefer the high-level task tools:

- `a_stock_investigation` for broad natural-language A-share investigation requests.
- `a_stock_quote_lookup` for realtime quote, K-line,盘口, turnover, PE/PB, market cap, and limit prices.
- `a_stock_research_lookup` for research reports, consensus forecasts, rating, and report PDF metadata.
- `a_stock_signal_lookup` for hot topics, reason tags, fund flow, northbound flow, Dragon Tiger Board, lockup expiry, and industry rotation.
- `a_stock_announcement_lookup` for CNINFO/F10 announcement lookup.
- `a_stock_profile_lookup` for company profile, industry, listing date, share capital, F10 profile, and basic quarterly snapshot.

## Grounding Rules

- Treat `entity_name`, `stock_code`, `market`, requested fields, and dates as candidate intent until a source adapter verifies them.
- For A-share price/code requests where the named subject may be private, overseas, or otherwise unsupported, still use `a_stock_quote_lookup` or `a_stock_investigation` so the adapter can return a structured `unsupported` / `identity_not_found` boundary; do not answer solely from model memory.
- Treat `identity_not_found` as an A-share resolver boundary only: say no verified A-share identity or code was found; do not infer private-company, global-listing, legal, or other-market status without separate evidence.
- Preserve the user's freshness request, such as realtime, today, latest trading day, last 20 days, future 90 days, or explicit date.
- Final answers must cite source, timestamp/as-of date, and whether data is realtime, intraday, EOD, historical, or cached.
- Final answers that include price, PE/PB, market cap, research/rating, signal, risk, valuation, or comparison evidence must include a short non-personalized investment-advice boundary.
- For valuation or investment-style outputs, separate source facts from formula-derived observations and risk boundaries.
- Quote/valuation snapshot tools provide point-in-time PE/PB/market-cap evidence only. Do not claim historical valuation percentile, cheap/expensive versus history, or peer-relative ranking unless a configured source explicitly provides that evidence.
- If a provider is unavailable, rate-limited, unconfigured, or stale, state that explicitly instead of fabricating missing fields.
- Broad stock screening is only supported when the host provides a verified screening adapter or workflow. Without that, use `a_stock_investigation` only for bounded multi-stock comparison over user-provided entities, and state the boundary.

## Boundary

This skill is for A-share market-data and signal workflows. For public annual reports, report metrics, full-report briefs, and financial statement extraction, prefer `agentx_finance` tools such as `finance_report_lookup`.

When `agentx_finance` is also loaded:

- Use `finance_report_lookup` for annual/quarterly report facts, revenue, net profit, growth, cash flow, report brief, and report-derived performance assessment.
- Use A-share tools for realtime or latest-trading-day quote, PE/PB, market cap, turnover, research reports, ratings, announcement lists, company profile, hot reasons, subject concept evidence, subject fund-flow evidence, Dragon Tiger Board evidence, lockup-expiry evidence, industry-board comparison evidence, and northbound-flow evidence when the host has configured the northbound host-cache adapter. Northbound flow must degrade when that host cache is unavailable or unconfigured.
- For mixed investment-style questions, gather report facts with `finance_report_lookup` first, then add A-share quote/research/signal evidence only if the user asks for valuation, market performance, institution views, or trading signals.
- Do not use A-share market-data tools to infer report metrics, and do not use finance report tools to fabricate realtime market data or trading-signal evidence.
- If a signal provider is unconfigured or unsupported, state that boundary explicitly rather than substituting report facts.
