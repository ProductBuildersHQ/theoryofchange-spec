# Credits

This specification stands on a long body of work on Theory of Change and on
evaluation. Where it takes an idea, a term, or a structure from that work, the
source is named here. The [research brief](research/theory-of-change.md) is the
full record: it summarizes each source and gives the citations.

## Methods and where they appear

| Source | What the spec takes from it | Where it appears |
|--------|-----------------------------|------------------|
| ActKnowledge and the Center for Theory of Change (Taplin and Clark) | Backwards mapping, preconditions, rationales for each link, the accountability ceiling, and the components of an indicator | Precondition groups, link `rationale`, `ceilingOfAccountability`, indicators |
| UNDG, *UNDAF Companion Guidance: Theory of Change* (2017) | The assumption types causality, implementation, external factors, and partnerships, the idea that each assumption is judged for plausibility and for whether it needs testing, and the problem tree of immediate, underlying, and structural causes | `Assumption.type`, `needsTesting`, `Cause.level`, the assumption lint rules |
| Organizational Research Services for the Annie E. Casey Foundation (Reisman and Gienapp, 2004) | Outcomes arranged along time horizons | Outcome levels and `lag` |
| Aspen Institute Roundtable on Community Change (Weiss, 1995; Anderson, 2005) | Theory-driven evaluation, and the three quality criteria of plausible, feasible, and testable | `qualityReview` on revisions and the review rubric |
| *Outcome Mapping* (Earl, Carden, and Smutylo, IDRC, 2001) | Spheres of control, influence, and interest, and boundary partners | `Outcome.sphere`, the `boundary_partner` actor role |
| Contribution analysis (Mayne) | Separating a claim of attribution from a claim of contribution | `Indicator.claim`, and the lint rule that indicators above the ceiling claim contribution |
| Process tracing (Van Evera, 1997; combined with contribution analysis by Befani and Mayne, 2014) | Grading evidence for a causal link with the straw-in-the-wind, hoop, smoking-gun, and doubly decisive tests | `Evidence.processTracingTest` |
| *Most Significant Change* (Davies and Dart, 2005) | Stories of change coded into themes so they can be counted | Stories, themes, and the story codebook |
| *Outcome Harvesting* (Wilson-Grau and Britt, 2012) | Outcomes found after the fact, as a kind of evidence | The `outcome_harvest` evidence type |
| The logic model and logical framework traditions (W.K. Kellogg Foundation, 2004; USAID; UNDG, *Results-Based Management Handbook*, 2011) | The linear views of a theory | The logic-model and logframe reports |
| Problem-tree analysis in the logical framework approach (GTZ, *ZOPP*, 1997; European Commission, *Project Cycle Management Guidelines*, 2004) | The core problem in the middle, its causes below as roots, and its effects above as branches | `effects`, `Cause.parents`, `Effect.parents`, `addressesEffects`, the problem-tree report |
| V2MOM (Marc Benioff, Salesforce; Benioff and Adler, 2009) | Vision, values, methods, obstacles, and measures, in that order, as one alignment document | `context.values` and the V2MOM export |
| Six Sigma's DMAIC cycle (Motorola; General Electric) | Improvement efforts that define, measure, analyze, improve, and control a metric from a baseline toward a target | `pathway.projections` with `dmaic`, the DMAIC export, and the `IMP-DMA` lint rules |
| Statistical process control (Shewhart, 1931) | Control limits from an individuals and moving-range chart, three sigma either side of the center line | Control limits in the DMAIC export |

Isabel Vogel's review of how Theory of Change is used in international development
(DFID, 2012) makes the point, quoted in the research brief, that a logic model
without reference to its context is closer to a logframe than to a theory of
change. In that spirit this spec keeps assumptions, risks, and context in the same
document as the pathways.

## Projects and software

This project is part of the PRISM family of projects by the same author. It uses
[prism-core](https://github.com/grokify/prism-core) for shared types, the goals
packages of [prism-roadmap](https://github.com/grokify/prism-roadmap) for the
OKR, outcome-journey, V2MOM, and DMAIC exports,
[d2vision](https://github.com/grokify/d2vision) for the layered logic-model and
problem-tree diagrams, and
[structured-evaluation](https://github.com/plexusone/structured-evaluation) for
rubrics and GO and NO-GO reports. Operating-system differences are handled by
[oscompat](https://github.com/grokify/oscompat).

The command line is built on [Cobra](https://github.com/spf13/cobra) and
[pflag](https://github.com/spf13/pflag), YAML is read and written with
[yaml.v3](https://github.com/go-yaml/yaml), documents are checked with
[santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema), the
schemas are generated from the Go types with
[invopop/jsonschema](https://github.com/invopop/jsonschema), and the MCP server
uses the [official Go SDK](https://github.com/modelcontextprotocol/go-sdk). The
list of every module compiled into the binary, with its license and copyright
notice, is in
[THIRD_PARTY_NOTICES.md](https://github.com/ProductBuildersHQ/impact-model/blob/main/THIRD_PARTY_NOTICES.md).

Some commands call [Marp](https://marp.app), the
[Mermaid CLI](https://github.com/mermaid-js/mermaid-cli), and
[Pandoc](https://pandoc.org) when they are installed. They are not included. This
site is built with [MkDocs](https://www.mkdocs.org) and the
[Material theme](https://squidfunk.github.io/mkdocs-material/), and its page
template comes from [visionspec](https://github.com/ProductBuildersHQ/visionspec).

## How the sources are used

- The definitions, guidance, and examples in this documentation are written in
  our own words. The research brief quotes sources in short passages, each
  attributed to its author.
- All examples are fictional: the organizations, people, numbers, and URLs are
  invented. No example describes a real organization.
- The names of organizations and publications are used to credit their work.
  Naming a source does not mean that its authors or publishers have reviewed or
  endorse this project.
- This project is licensed under the [MIT License](https://github.com/ProductBuildersHQ/impact-model/blob/main/LICENSE). The software it depends on keeps its own
  licenses, which are all permissive.

If a credit is missing or wrong, please
[open an issue](https://github.com/ProductBuildersHQ/impact-model/issues).

## Works cited above

- Taplin, D. and Clark, H. (2012). *Theory of Change Basics: A Primer on Theory of Change*. ActKnowledge. <https://www.theoryofchange.org/wp-content/uploads/toco_library/pdf/ToCBasics.pdf>
- Taplin, D., Clark, H., Collins, E. and Colby, D. (2013). *Theory of Change Technical Papers*. ActKnowledge. <http://www.actknowledge.org/resources/documents/ToC-Tech-Papers.pdf>
- UNDG (2017). *UNDAF Companion Guidance: Theory of Change*. <https://unsdg.un.org/sites/default/files/UNDG-UNDAF-Companion-Pieces-7-Theory-of-Change.pdf>
- UNDG (2011). *Results-Based Management Handbook*. <https://unsdg.un.org/resources/results-based-management-handbook>
- Reisman, J. and Gienapp, A. (2004). *Theory of Change: A Practical Tool for Action, Results and Learning*. Organizational Research Services for the Annie E. Casey Foundation. <https://www.aecf.org/resources/theory-of-change>
- Weiss, C. (1995). "Nothing as Practical as Good Theory." In Connell, Kubisch, Schorr and Weiss (eds.), *New Approaches to Evaluating Comprehensive Community Initiatives*. Aspen Institute.
- Anderson, A. (2005). *The Community Builder's Approach to Theory of Change*. Aspen Institute Roundtable on Community Change.
- Earl, S., Carden, F. and Smutylo, T. (2001). *Outcome Mapping: Building Learning and Reflection into Development Programs*. International Development Research Centre.
- Mayne, J. Contribution analysis (2001, 2008, 2012). Overview: <https://betterevaluation.org/node/2708>
- Mayne, J. (2017). "Theory of change analysis: Building robust theories of change." *Canadian Journal of Program Evaluation*. <https://doi.org/10.3138/cjpe.31122>
- Van Evera, S. (1997). *Guide to Methods for Students of Political Science*. Cornell University Press.
- Befani, B. and Mayne, J. (2014). "Process Tracing and Contribution Analysis: A Combined Approach to Generative Causal Inference." *IDS Bulletin* 45(6).
- Davies, R. and Dart, J. (2005). *The 'Most Significant Change' (MSC) Technique: A Guide to Its Use*. <https://mande.co.uk/>
- Wilson-Grau, R. and Britt, H. (2012). *Outcome Harvesting*. Ford Foundation. <https://outcomeharvesting.net/>
- W.K. Kellogg Foundation (2004). *Logic Model Development Guide*.
- GTZ (1997). *ZOPP: Objectives-oriented Project Planning*. Deutsche Gesellschaft für Technische Zusammenarbeit.
- European Commission (2004). *Aid Delivery Methods, Volume 1: Project Cycle Management Guidelines*. EuropeAid.
- Benioff, M. and Adler, C. (2009). *Behind the Cloud*. Jossey-Bass.
- Shewhart, W. A. (1931). *Economic Control of Quality of Manufactured Product*. Van Nostrand.
- Vogel, I. (2012). *Review of the use of 'Theory of Change' in International Development*. DFID. <https://assets.publishing.service.gov.uk/media/57a08a5ded915d3cfd00071a/DFID_ToC_Review_VogelV7.pdf>
