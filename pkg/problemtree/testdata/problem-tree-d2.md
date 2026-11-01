## Problem tree

Causes (roots) below lead up to the core problem, which leads up to its effects (branches).

```d2
grid-columns: 1
vertical-gap: 70
direction: up

classes: {
  cause: {
    style.fill: "#f3e6d8"
    style.stroke: "#8d5a2b"
    style.border-radius: 12
  }
  effect: {
    style.fill: "#e4f3e5"
    style.stroke: "#2e7d32"
    style.border-radius: 12
  }
  problem: {
    style.fill: "#ffe0b2"
    style.stroke: "#e08a1e"
    style.border-radius: 12
    style.font-size: 20
    style.bold: true
  }
  row: {
    style.fill: "transparent"
    style.stroke-width: 0
  }
  set_aside: {
    style.fill: "#f3f3f3"
    style.stroke: "#9e9e9e"
    style.border-radius: 12
    style.opacity: 0.70
  }
  zone: {
    style.fill: "transparent"
    style.stroke-width: 0
    style.font-size: 20
    style.bold: true
  }
}

effects: Effects (branches) {
  grid-columns: 1
  vertical-gap: 70
  class: zone

  r2: "" {
    grid-rows: 1
    horizontal-gap: 30
    class: row

    "eff-malnutrition": Malnutrition and low energy {
      class: effect
    }
    "eff-costs": Higher health costs {
      class: effect
    }
    "eff-dental": Increased dental care {
      class: effect
    }
  }
  r1: "" {
    grid-rows: 1
    horizontal-gap: 30
    class: row

    "eff-weight": Overweight or underweight {
      class: effect
    }
    "eff-diabetes": Diabetes and heart disease {
      class: effect
    }
    "eff-decay": Tooth decay {
      class: effect
    }
    "eff-mental": Depression, anxiety, and eating disorders {
      class: effect
    }
  }
}

core: Core problem {
  grid-columns: 1
  vertical-gap: 70
  class: zone

  r0: "" {
    grid-rows: 1
    horizontal-gap: 30
    class: row

    problem: Poor eating health {
      class: problem
    }
  }
}

causes: Causes (roots) {
  grid-columns: 1
  vertical-gap: 70
  class: zone

  r_1: "" {
    grid-rows: 1
    horizontal-gap: 30
    class: row

    "cause-stress": Stress {
      class: cause
    }
    "cause-busy": Too busy {
      class: cause
    }
    "cause-poverty": Poverty (set aside) {
      class: set_aside
    }
    "cause-water": Low water intake {
      class: cause
    }
  }
  r_2: "" {
    grid-rows: 1
    horizontal-gap: 30
    class: row

    "cause-fastfood": Availability of fast food {
      class: cause
    }
    "cause-alcohol": High alcohol intake {
      class: cause
    }
  }
}

causes.r_1."cause-stress" -> core.r0.problem
causes.r_1."cause-busy" -> core.r0.problem
causes.r_2."cause-fastfood" -> causes.r_1."cause-busy"
causes.r_1."cause-poverty" -> core.r0.problem
causes.r_1."cause-water" -> core.r0.problem
causes.r_2."cause-alcohol" -> causes.r_1."cause-water"
core.r0.problem -> effects.r1."eff-weight"
effects.r1."eff-weight" -> effects.r2."eff-malnutrition"
core.r0.problem -> effects.r1."eff-diabetes"
effects.r1."eff-diabetes" -> effects.r2."eff-costs"
core.r0.problem -> effects.r1."eff-decay"
effects.r1."eff-decay" -> effects.r2."eff-dental"
core.r0.problem -> effects.r1."eff-mental"
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

