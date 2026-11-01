---
title: "Open agent authorization standard adoption"
version: "0.1"
model-id: "open-agent-auth"
content-hash: "6c01012150da0627ca84b9b55e045a952149af0a1ea2f8ee02727bfeed5866c1"
---

# The problem

AI agents act on behalf of users with credentials designed for humans, so access is over-broad and unauditable.

**Root causes we see**

- No standard way to express delegated agent authority

::: notes
Ask: would you describe the problem this way? Who is missing from this description?
:::

---

# The change we are working toward

- **Agent access across the market is scoped and auditable**

**Ceiling of accountability:** Long-term outcome and below

Market-wide security outcomes depend on many actors.

::: notes
Outcomes above the ceiling are tracked but not owned: we contribute to them, we do not claim them. Ask: is the line in the right place?
:::

---

# The theory of change on one page

![Diagram](d.png)

::: notes
Walk the arrows upward. Ask: is each arrow really a cause? What is missing between a box and the one above it?
:::

---

# Pathway: Supply side: providers and platforms

![Diagram](d.png)

::: notes
Ask the people who run this pathway to correct it. What do you do that is not on the page? What on the page do you not do?
:::

---

# Logic model: Supply side: providers and platforms

| Stage | Elements |
|-------|----------|
| Inputs | Two engineers and an event budget |
| Activities | Run interoperability events; Publish reference implementations |
| Outputs | Interop test suite; Event reports; Open reference implementations are available |
| Short-term outcomes | A stable, interoperable specification exists |
| Medium-term outcomes | Major identity providers support the standard [inferred]; Agent platforms support the standard |
| Long-term outcomes | Enterprises require the standard in procurement [hypothesis] |
| Impact | Agent access across the market is scoped and auditable |

::: notes
Read left to right: what goes in, what we do, what we produce, what changes. A blank row is a gap worth discussing.
:::

---

# Logic model, step by step: Supply side: providers and platforms

![Diagram](d.png)

::: notes
Follow one input to the change it is for. Is each arrow a thing that really leads to the next? Where does an arrow have no reason behind it?
:::

---

# What we are assuming

**Identity providers see commercial advantage in supporting agent delegation.** [inferred]

- If the specification is stable, then providers implement it, because customers are asking for agent controls
- To test it: Interview five providers about roadmap priorities.
- Criticality high · status Untested

**Large vendors do not fork the standard.**

- Criticality critical · status Holding

::: notes
Ask: which of these do you believe least? What evidence would change your mind?
:::

---

# How we will measure

| Metric | Definition | Formula | Combines across pathways |
|--------|------------|---------|--------------------------|
| Providers with GA support | Identity providers shipping the standard in a general-availability release. | Count of distinct providers with a GA release implementing the standard. | Sum |
| Share of provider seats covered | Share of provider seats that can use the standard. | Seats at providers with GA support divided by seats at all tracked providers, weighted by seats. | Mean weighted by Providers with GA support |

::: notes
Ask: could you collect this with the time you have? Is the formula what you would count?
:::

---

# Words we are using

| Term | Meaning here |
|------|--------------|
| General availability | A release available to all customers without special arrangement, as opposed to a preview. |

::: notes
Agree the words before you debate the content. Which of these would your funders define differently?
:::

---

# What is established, and what is our reading

| Basis | Elements |
|-------|----------|
| Stated | 3 |
| Inferred | 2 |
| Hypothesis | 2 |
| Confirmed | 4 |
| No basis recorded | 19 |

::: notes
Stated means a source says it. Inferred and hypothesis are our reading and are the first things to challenge. Confirmed means the people it concerns agreed.
:::

---

# For the room to settle

**Our reading, to confirm or correct**

- Enterprises require the standard in procurement (hypothesis)
- Major identity providers support the standard (inferred)
- Identity providers see commercial advantage in supporting agent delegation. (inferred)

::: notes
Work down this list. For each: agree, change, remove, or not sure. Record what the room decides.
:::

---

# Export details

Impact model open-agent-auth · version 0.1 · hash 6c01012150da
