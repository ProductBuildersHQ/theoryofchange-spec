# Open agent authorization standard adoption: review packet

## How to review

This is a draft for you to correct. For each row, write one of agree, change, remove, unsure in the **Decision** column. If you choose change, say what it should be in **Suggested change**. Use **Comment** for anything else. Please leave the ID column as it is: it is how your answers are matched to the theory.

Rows marked inferred or hypothesis are our reading, not something the organization said, and are the most important to check.

## Outcomes

| ID | Statement | Basis | Sources | Decision | Suggested change | Comment |
|----|-----------|-------|---------|----------|------------------|---------|
| oc-safer-agents | Impact: Agent access across the market is scoped and auditable | — | — |  |  |  |
| oc-enterprise-require | Long-term outcome: Enterprises require the standard in procurement | hypothesis | — |  |  |  |
| oc-idp-support | Intermediate outcome: Major identity providers support the standard | inferred | ev-survey: Industry survey on agent access controls |  |  |  |
| oc-platform-support | Intermediate outcome: Agent platforms support the standard | — | — |  |  |  |
| oc-spec-stable | Short-term outcome: A stable, interoperable specification exists | stated | ev-survey: Industry survey on agent access controls |  |  |  |
| oc-reference-impl | Output: Open reference implementations are available | — | — |  |  |  |

## Causal links

| ID | Statement | Basis | Sources | Decision | Suggested change | Comment |
|----|-----------|-------|---------|----------|------------------|---------|
| ln-spec-idp | A stable, interoperable specification exists → Major identity providers support the standard. Providers implement only stable specifications. | stated | ev-survey: Industry survey on agent access controls |  |  |  |
| ln-impl-platform | Open reference implementations are available → Agent platforms support the standard. Reference code lowers integration cost. | — | — |  |  |  |
| ln-idp-require | Major identity providers support the standard → Enterprises require the standard in procurement | — | — |  |  |  |
| ln-platform-require | Agent platforms support the standard → Enterprises require the standard in procurement | — | — |  |  |  |
| ln-require-safer | Enterprises require the standard in procurement → Agent access across the market is scoped and auditable. Procurement requirements drive market-wide behavior. | — | — |  |  |  |
| ln-require-platform-loop | Enterprises require the standard in procurement → Agent platforms support the standard | — | — |  |  |  |
| grp-supply | Precondition group (all of its members) leading to Enterprises require the standard in procurement | confirmed | — |  |  |  |

## Assumptions

| ID | Statement | Basis | Sources | Decision | Suggested change | Comment |
|----|-----------|-------|---------|----------|------------------|---------|
| asm-idp-incentive | Identity providers see commercial advantage in supporting agent delegation. (If the specification is stable, then providers implement it, because customers are asking for agent controls) | inferred | ev-survey: Industry survey on agent access controls |  |  |  |
| asm-no-fork | Large vendors do not fork the standard. | — | — |  |  |  |

## Interventions

| ID | Statement | Basis | Sources | Decision | Suggested change | Comment |
|----|-----------|-------|---------|----------|------------------|---------|
| int-interop | Run interoperability events | confirmed | — |  |  |  |
| int-reference | Publish reference implementations | — | — |  |  |  |

## Indicators

| ID | Statement | Basis | Sources | Decision | Suggested change | Comment |
|----|-----------|-------|---------|----------|------------------|---------|
| ind-idp-count | Identity providers with generally available support, measuring Major identity providers support the standard | hypothesis | — |  |  |  |
| ind-idp-roadmaps | Providers with the standard on their public roadmap, measuring asm-idp-incentive | — | — |  |  |  |
| ind-reference-downloads | Reference implementation downloads, measuring Publish reference implementations | — | — |  |  |  |
| ind-idp-count-na | Providers with GA support (North America), measuring Major identity providers support the standard | — | — |  |  |  |
| ind-idp-count-eu | Providers with GA support (Europe), measuring Major identity providers support the standard | — | — |  |  |  |
| ind-idp-total | Providers with GA support (total), measuring Major identity providers support the standard | — | — |  |  |  |
| ind-adoption-stories | Stories of provider adoption, measuring Agent access across the market is scoped and auditable | — | — |  |  |  |

## Metrics

| ID | Statement | Basis | Sources | Decision | Suggested change | Comment |
|----|-----------|-------|---------|----------|------------------|---------|
| met-providers | Providers with GA support: Identity providers shipping the standard in a general-availability release. | confirmed | — |  |  |  |
| met-provider-share | Share of provider seats covered: Share of provider seats that can use the standard. | — | — |  |  |  |

## Actors

| ID | Statement | Basis | Sources | Decision | Suggested change | Comment |
|----|-----------|-------|---------|----------|------------------|---------|
| act-proponent | Standards working group (Group) | — | — |  |  |  |
| act-idp | Identity providers (Vendor) | stated | ev-survey: Industry survey on agent access controls |  |  |  |
| act-enterprises | Enterprise security buyers (Customer segment) | — | — |  |  |  |

