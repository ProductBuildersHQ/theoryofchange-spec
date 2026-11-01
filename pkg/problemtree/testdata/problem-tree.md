## Problem tree

Causes (roots) below lead up to the core problem, which leads up to its effects (branches).

```mermaid
flowchart BT
  classDef cause fill:#f3e6d8,stroke:#8d5a2b
  classDef effect fill:#e4f3e5,stroke:#2e7d32
  classDef problem fill:#ffe0b2,stroke:#e08a1e
  classDef row fill:transparent,stroke-width:0px
  classDef set_aside fill:#f3f3f3,stroke:#9e9e9e
  classDef zone fill:transparent,stroke-width:0px
  subgraph n_effects["Effects (branches)"]
    subgraph n_effects_r2[""]
      n_effects_r2_eff_malnutrition["Malnutrition and low energy"]
      n_effects_r2_eff_costs["Higher health costs"]
      n_effects_r2_eff_dental["Increased dental care"]
    end
    subgraph n_effects_r1[""]
      n_effects_r1_eff_weight["Overweight or underweight"]
      n_effects_r1_eff_diabetes["Diabetes and heart disease"]
      n_effects_r1_eff_decay["Tooth decay"]
      n_effects_r1_eff_mental["Depression, anxiety, and eating disorders"]
    end
  end
  subgraph n_core["Core problem"]
    subgraph n_core_r0[""]
      n_core_r0_problem["Poor eating health"]
    end
  end
  subgraph n_causes["Causes (roots)"]
    subgraph n_causes_r_1[""]
      n_causes_r_1_cause_stress["Stress"]
      n_causes_r_1_cause_busy["Too busy"]
      n_causes_r_1_cause_poverty["Poverty (set aside)"]
      n_causes_r_1_cause_water["Low water intake"]
    end
    subgraph n_causes_r_2[""]
      n_causes_r_2_cause_fastfood["Availability of fast food"]
      n_causes_r_2_cause_alcohol["High alcohol intake"]
    end
  end
  n_causes_r_1_cause_stress --> n_core_r0_problem
  n_causes_r_1_cause_busy --> n_core_r0_problem
  n_causes_r_2_cause_fastfood --> n_causes_r_1_cause_busy
  n_causes_r_1_cause_poverty --> n_core_r0_problem
  n_causes_r_1_cause_water --> n_core_r0_problem
  n_causes_r_2_cause_alcohol --> n_causes_r_1_cause_water
  n_core_r0_problem --> n_effects_r1_eff_weight
  n_effects_r1_eff_weight --> n_effects_r2_eff_malnutrition
  n_core_r0_problem --> n_effects_r1_eff_diabetes
  n_effects_r1_eff_diabetes --> n_effects_r2_eff_costs
  n_core_r0_problem --> n_effects_r1_eff_decay
  n_effects_r1_eff_decay --> n_effects_r2_eff_dental
  n_core_r0_problem --> n_effects_r1_eff_mental
  class n_effects_r2_eff_malnutrition effect
  class n_effects_r2_eff_costs effect
  class n_effects_r2_eff_dental effect
  class n_effects_r2 row
  class n_effects_r1_eff_weight effect
  class n_effects_r1_eff_diabetes effect
  class n_effects_r1_eff_decay effect
  class n_effects_r1_eff_mental effect
  class n_effects_r1 row
  class n_effects zone
  class n_core_r0_problem problem
  class n_core_r0 row
  class n_core zone
  class n_causes_r_1_cause_stress cause
  class n_causes_r_1_cause_busy cause
  class n_causes_r_1_cause_poverty set_aside
  class n_causes_r_1_cause_water cause
  class n_causes_r_1 row
  class n_causes_r_2_cause_fastfood cause
  class n_causes_r_2_cause_alcohol cause
  class n_causes_r_2 row
  class n_causes zone
```

### Causes

| Cause | Level | Leads to | Addressed by |
|-------|-------|----------|--------------|
| Stress (`cause-stress`) | immediate | the core problem | — |
| Too busy (`cause-busy`) | immediate | the core problem | People have time to cook (`oc-time`) |
| Availability of fast food (`cause-fastfood`) | underlying | Too busy (`cause-busy`) | — |
| Poverty (`cause-poverty`) | structural | the core problem | set aside: Beyond this program. |
| Low water intake (`cause-water`) | immediate | the core problem | — |
| High alcohol intake (`cause-alcohol`) | underlying | Low water intake (`cause-water`) | — |

### Effects

| Effect | Follows from | Relieved by |
|--------|--------------|-------------|
| Overweight or underweight (`eff-weight`) | the core problem | — |
| Malnutrition and low energy (`eff-malnutrition`) | Overweight or underweight (`eff-weight`) | — |
| Diabetes and heart disease (`eff-diabetes`) | the core problem | People eat well (`oc-healthy`) |
| Higher health costs (`eff-costs`) | Diabetes and heart disease (`eff-diabetes`) | — |
| Tooth decay (`eff-decay`) | the core problem | — |
| Increased dental care (`eff-dental`) | Tooth decay (`eff-decay`) | — |
| Depression, anxiety, and eating disorders (`eff-mental`) | the core problem | — |

