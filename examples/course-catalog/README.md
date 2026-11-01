# Course catalog example (fictional)

Everything here is **fictional**: Lanternwood Learning Commons, Brookfield
Volunteer Network, every course, all numbers, and all `example.org` URLs. It is
invented to show the format and resembles no real organization or program.

A small nonprofit publishes free, on-demand certificate courses for staff of
other nonprofits. Starting a course is easy and finishing one is not, so the
catalog program is run as a DMAIC effort on its weekly completion rate, with a
post-course Net Promoter Score as the guardrail.

## Files

- `impact-model.yaml`: the impact model
- `observations-2026-h1.yaml`: 24 weeks of observations for the weekly indicators

## Features shown

- A pathway marked `projections: [dmaic]`, exported with `export dmaic`
- Weekly indicators measured from a platform, a survey, and a partner's forum
- A survey indicator with enough respondents for each weekly measurement
- Indicators the DMAIC lint rules flag on purpose: a catalog count taken
  monthly, a graduate survey taken once a year, and a yearly site-visit rating
  with no data source
- A feedback loop: graduates who apply what they learned recommend the catalog
- Root causes in the problem tree that the completion and enrollment outcomes
  address
- Values, so the model also exports as a V2MOM

## Try it

```bash
impact-model lint impact-model.yaml
impact-model export dmaic impact-model.yaml -O observations-2026-h1.yaml -o dmaic.json
impact-model export v2mom impact-model.yaml -O observations-2026-h1.yaml -o v2mom.json
```
