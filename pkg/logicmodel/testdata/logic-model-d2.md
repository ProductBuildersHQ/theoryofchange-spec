## Layered logic model: Learning Exchange

```d2
grid-columns: 7
horizontal-gap: 90
direction: right

classes: {
  activities: {
    style.fill: "#e4f3e5"
    style.stroke: "#2e7d32"
    style.border-radius: 12
  }
  impact: {
    style.fill: "#c5cfe8"
    style.stroke: "#1f3a7a"
    style.border-radius: 12
  }
  inputs: {
    style.fill: "#fff1dc"
    style.stroke: "#e08a1e"
    style.border-radius: 12
  }
  long_term: {
    style.fill: "#d3dcee"
    style.stroke: "#2c4a8a"
    style.border-radius: 12
  }
  medium_term: {
    style.fill: "#dfe8f6"
    style.stroke: "#3f5fa8"
    style.border-radius: 12
  }
  outputs: {
    style.fill: "#ece4fb"
    style.stroke: "#7e57c2"
    style.border-radius: 12
  }
  short_term: {
    style.fill: "#e1f3fd"
    style.stroke: "#2d9cdb"
    style.border-radius: 12
  }
  stage: {
    style.fill: "transparent"
    style.stroke-width: 0
    style.font-size: 22
    style.bold: true
  }
}

inputs: Inputs {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  _input1: Two program staff and meeting space donated by the county library system {
    class: inputs
  }
  _input2: One program staff member and a pool of volunteer coaches {
    class: inputs
  }
  _input3: Fee-for-service agreement with the training partner {
    class: inputs
  }
}

activities: Activities {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "int-peer-cohorts": Run mixed-size peer-learning cohorts {
    class: activities
  }
  "int-leader-coaching": Offer one-to-one leadership coaching {
    class: activities
  }
  "int-training-series": Deliver a management skills training series (inferred) {
    class: activities
  }
}

outputs: Outputs {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  _output1: Cohort sessions held {
    class: outputs
  }
  _output2: Peer-pairing introductions {
    class: outputs
  }
  "oc-out-cohorts": Peer-learning cohorts and coaching sessions delivered {
    class: outputs
  }
  _output3: Coaching hours delivered {
    class: outputs
  }
  _output4: Workshops delivered {
    class: outputs
  }
  _output5: Participant materials {
    class: outputs
  }
}

short_term: Short-term outcomes {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-st-peer-ties": Members form peer relationships across sizes (inferred) {
    class: short_term
  }
  "oc-st-leader-skills": Leaders gain coaching-based skills and confidence {
    class: short_term
  }
}

medium_term: Medium-term outcomes {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-im-peer-network": Members turn to each other for advice and joint work (inferred) {
    class: medium_term
  }
  "oc-im-joint-positions": Members adopt and share joint positions on local issues (hypothesis) {
    class: medium_term
  }
  "oc-im-leaders-practice": Member leaders apply improved management practices {
    class: medium_term
  }
}

long_term: Long-term outcomes {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-lt-collab": Members routinely collaborate instead of duplicating effort {
    class: long_term
  }
  "oc-lt-collective-voice": Local government and funders consult members through a coordinated channel (hypothesis) {
    class: long_term
  }
  "oc-lt-capable-members": Member nonprofits run strong, well-managed organizations {
    class: long_term
  }
}

impact: Impact {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-resident-voice": 'Residents'' priorities shape local public decisions (hypothesis)' {
    class: impact
  }
  "oc-resident-access": Residents get timely, good-quality services from local nonprofits (inferred) {
    class: impact
  }
}

inputs._input1 -> activities."int-peer-cohorts"
activities."int-peer-cohorts" -> outputs._output1
activities."int-peer-cohorts" -> outputs._output2
activities."int-peer-cohorts" -> outputs."oc-out-cohorts"
outputs._output1 -> short_term."oc-st-peer-ties"
outputs._output2 -> short_term."oc-st-peer-ties"
inputs._input2 -> activities."int-leader-coaching"
activities."int-leader-coaching" -> outputs._output3
activities."int-leader-coaching" -> outputs."oc-out-cohorts"
outputs._output3 -> short_term."oc-st-leader-skills"
inputs._input3 -> activities."int-training-series"
activities."int-training-series" -> outputs._output4
activities."int-training-series" -> outputs._output5
outputs._output4 -> short_term."oc-st-leader-skills"
outputs._output5 -> short_term."oc-st-leader-skills"
outputs."oc-out-cohorts" -> short_term."oc-st-leader-skills"
outputs."oc-out-cohorts" -> short_term."oc-st-peer-ties"
short_term."oc-st-leader-skills" -> medium_term."oc-im-leaders-practice"
short_term."oc-st-peer-ties" -> medium_term."oc-im-peer-network"
medium_term."oc-im-leaders-practice" -> long_term."oc-lt-capable-members"
medium_term."oc-im-peer-network" -> long_term."oc-lt-collab"
short_term."oc-st-peer-ties" -> medium_term."oc-im-joint-positions"
medium_term."oc-im-joint-positions" -> long_term."oc-lt-collab"
medium_term."oc-im-joint-positions" -> long_term."oc-lt-collective-voice"
long_term."oc-lt-capable-members" -> impact."oc-resident-access"
long_term."oc-lt-collab" -> impact."oc-resident-access"
long_term."oc-lt-collective-voice" -> impact."oc-resident-voice"
long_term."oc-lt-collab" -> impact."oc-resident-voice"
```

Feedback, not drawn:

- Members routinely collaborate instead of duplicating effort leads back to Members form peer relationships across sizes (inferred)

## Layered logic model: Shared Services

```d2
grid-columns: 7
horizontal-gap: 90
direction: right

classes: {
  activities: {
    style.fill: "#e4f3e5"
    style.stroke: "#2e7d32"
    style.border-radius: 12
  }
  impact: {
    style.fill: "#c5cfe8"
    style.stroke: "#1f3a7a"
    style.border-radius: 12
  }
  inputs: {
    style.fill: "#fff1dc"
    style.stroke: "#e08a1e"
    style.border-radius: 12
  }
  long_term: {
    style.fill: "#d3dcee"
    style.stroke: "#2c4a8a"
    style.border-radius: 12
  }
  medium_term: {
    style.fill: "#dfe8f6"
    style.stroke: "#3f5fa8"
    style.border-radius: 12
  }
  outputs: {
    style.fill: "#ece4fb"
    style.stroke: "#7e57c2"
    style.border-radius: 12
  }
  short_term: {
    style.fill: "#e1f3fd"
    style.stroke: "#2d9cdb"
    style.border-radius: 12
  }
  stage: {
    style.fill: "transparent"
    style.stroke-width: 0
    style.font-size: 22
    style.bold: true
  }
}

inputs: Inputs {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  _input1: One shared-services coordinator and contracted bookkeepers {
    class: inputs
  }
  _input2: Part of one staff role {
    class: inputs
  }
}

activities: Activities {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "int-pooled-backoffice": Run pooled bookkeeping and payroll help {
    class: activities
  }
  "int-group-purchasing": Negotiate group purchasing agreements {
    class: activities
  }
  "int-grants-calendar": Maintain a shared grants calendar {
    class: activities
  }
}

outputs: Outputs {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  _output1: Member accounts onboarded {
    class: outputs
  }
  _output2: Shared service desk {
    class: outputs
  }
  "oc-out-services": Pooled service contracts and group purchasing deals in place {
    class: outputs
  }
  _output3: Vendor agreements signed {
    class: outputs
  }
  "oc-out-calendar": Shared grants calendar published and kept current (inferred) {
    class: outputs
  }
  _output4: Monthly calendar update {
    class: outputs
  }
  _output5: Deadline reminders {
    class: outputs
  }
}

short_term: Short-term outcomes {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-st-pooled-use": Members use pooled back-office services (inferred) {
    class: short_term
  }
  "oc-st-calendar-use": Members consult the shared grants calendar when planning (hypothesis) {
    class: short_term
  }
}

medium_term: Medium-term outcomes {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-im-lower-overhead": Members spend less staff time and money on back-office work {
    class: medium_term
  }
  "oc-im-grant-pipeline": Members apply for more grants that fit their mission (inferred) {
    class: medium_term
  }
}

long_term: Long-term outcomes {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-lt-capable-members": Member nonprofits run strong, well-managed organizations {
    class: long_term
  }
  "oc-lt-collab": Members routinely collaborate instead of duplicating effort {
    class: long_term
  }
  "oc-lt-sustainable": Members have diversified, stable funding (inferred) {
    class: long_term
  }
}

impact: Impact {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-resident-access": Residents get timely, good-quality services from local nonprofits (inferred) {
    class: impact
  }
  "oc-resident-voice": 'Residents'' priorities shape local public decisions (hypothesis)' {
    class: impact
  }
}

inputs._input1 -> activities."int-pooled-backoffice"
activities."int-pooled-backoffice" -> outputs._output1
activities."int-pooled-backoffice" -> outputs._output2
activities."int-pooled-backoffice" -> outputs."oc-out-services"
outputs._output1 -> short_term."oc-st-pooled-use"
outputs._output2 -> short_term."oc-st-pooled-use"
activities."int-group-purchasing" -> outputs._output3
activities."int-group-purchasing" -> outputs."oc-out-services"
inputs._input2 -> activities."int-grants-calendar"
activities."int-grants-calendar" -> outputs._output4
activities."int-grants-calendar" -> outputs._output5
activities."int-grants-calendar" -> outputs."oc-out-calendar"
outputs._output4 -> short_term."oc-st-calendar-use"
outputs._output5 -> short_term."oc-st-calendar-use"
outputs."oc-out-services" -> short_term."oc-st-pooled-use"
outputs."oc-out-calendar" -> short_term."oc-st-calendar-use"
short_term."oc-st-pooled-use" -> medium_term."oc-im-lower-overhead"
short_term."oc-st-calendar-use" -> medium_term."oc-im-grant-pipeline"
medium_term."oc-im-lower-overhead" -> long_term."oc-lt-capable-members"
medium_term."oc-im-lower-overhead" -> long_term."oc-lt-collab"
medium_term."oc-im-lower-overhead" -> long_term."oc-lt-sustainable"
medium_term."oc-im-grant-pipeline" -> long_term."oc-lt-sustainable"
long_term."oc-lt-capable-members" -> impact."oc-resident-access"
long_term."oc-lt-collab" -> impact."oc-resident-access"
long_term."oc-lt-sustainable" -> impact."oc-resident-access"
long_term."oc-lt-collab" -> impact."oc-resident-voice"
```

## Layered logic model: Civic Voice

```d2
grid-columns: 6
horizontal-gap: 90
direction: right

classes: {
  activities: {
    style.fill: "#e4f3e5"
    style.stroke: "#2e7d32"
    style.border-radius: 12
  }
  impact: {
    style.fill: "#c5cfe8"
    style.stroke: "#1f3a7a"
    style.border-radius: 12
  }
  long_term: {
    style.fill: "#d3dcee"
    style.stroke: "#2c4a8a"
    style.border-radius: 12
  }
  medium_term: {
    style.fill: "#dfe8f6"
    style.stroke: "#3f5fa8"
    style.border-radius: 12
  }
  outputs: {
    style.fill: "#ece4fb"
    style.stroke: "#7e57c2"
    style.border-radius: 12
  }
  short_term: {
    style.fill: "#e1f3fd"
    style.stroke: "#2d9cdb"
    style.border-radius: 12
  }
  stage: {
    style.fill: "transparent"
    style.stroke-width: 0
    style.font-size: 22
    style.bold: true
  }
}

activities: Activities {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "int-policy-briefings": Publish plain-language policy briefings (hypothesis) {
    class: activities
  }
  "int-joint-forums": Convene member forums to draft joint positions {
    class: activities
  }
  "int-liaison-channel": Set up a standing liaison channel with the county planning office {
    class: activities
  }
  "int-funder-roundtable": Host a funder and member roundtable {
    class: activities
  }
}

outputs: Outputs {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-out-briefings": Policy briefings and joint-position drafts produced (hypothesis) {
    class: outputs
  }
  _output1: Briefings published {
    class: outputs
  }
  _output2: Forums held {
    class: outputs
  }
  _output3: Joint position drafts {
    class: outputs
  }
  _output4: Quarterly liaison meeting {
    class: outputs
  }
  _output5: Named contacts on both sides {
    class: outputs
  }
  _output6: Roundtable held twice a year {
    class: outputs
  }
}

short_term: Short-term outcomes {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-st-policy-know": Members understand local policy and funding processes (inferred) {
    class: short_term
  }
  "oc-st-channel": A standing channel exists between members and the county planning office {
    class: short_term
  }
}

medium_term: Medium-term outcomes {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-im-joint-positions": Members adopt and share joint positions on local issues (hypothesis) {
    class: medium_term
  }
  "oc-im-officials-engage": Local officials and funders engage members in planning {
    class: medium_term
  }
}

long_term: Long-term outcomes {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-lt-collab": Members routinely collaborate instead of duplicating effort {
    class: long_term
  }
  "oc-lt-collective-voice": Local government and funders consult members through a coordinated channel (hypothesis) {
    class: long_term
  }
}

impact: Impact {
  grid-columns: 1
  vertical-gap: 24
  class: stage

  "oc-resident-access": Residents get timely, good-quality services from local nonprofits (inferred) {
    class: impact
  }
  "oc-resident-voice": 'Residents'' priorities shape local public decisions (hypothesis)' {
    class: impact
  }
}

activities."int-policy-briefings" -> outputs._output1
activities."int-policy-briefings" -> outputs."oc-out-briefings"
outputs._output1 -> short_term."oc-st-policy-know"
activities."int-joint-forums" -> outputs._output2
activities."int-joint-forums" -> outputs._output3
outputs._output2 -> short_term."oc-st-policy-know"
outputs._output3 -> short_term."oc-st-policy-know"
outputs._output2 -> medium_term."oc-im-joint-positions"
outputs._output3 -> medium_term."oc-im-joint-positions"
activities."int-liaison-channel" -> outputs._output4
activities."int-liaison-channel" -> outputs._output5
outputs._output4 -> short_term."oc-st-channel"
outputs._output5 -> short_term."oc-st-channel"
activities."int-funder-roundtable" -> outputs._output6
outputs._output6 -> short_term."oc-st-channel"
outputs._output6 -> medium_term."oc-im-officials-engage"
outputs."oc-out-briefings" -> short_term."oc-st-policy-know"
short_term."oc-st-policy-know" -> medium_term."oc-im-joint-positions"
medium_term."oc-im-joint-positions" -> long_term."oc-lt-collab"
medium_term."oc-im-joint-positions" -> long_term."oc-lt-collective-voice"
medium_term."oc-im-officials-engage" -> long_term."oc-lt-collective-voice"
short_term."oc-st-channel" -> medium_term."oc-im-officials-engage"
long_term."oc-lt-collab" -> impact."oc-resident-access"
long_term."oc-lt-collective-voice" -> impact."oc-resident-voice"
long_term."oc-lt-collab" -> impact."oc-resident-voice"
```

