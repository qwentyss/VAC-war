
# VAC Research

**Independent research and technical documentation of Valve Anti-Cheat (VAC) in Counter-Strike 2.**

Created on October 11, 2026, following reports of changes to cheat compatibility after the October 9, 2026 update.

This repository documents research, technical theories, observations, and tests related to VAC behavior before and after game updates.

> **Research status:** Ongoing. Hypotheses are not confirmed facts unless supported by reproducible evidence.

---

## Table of Contents

- [Research Goals](#research-goals)
- [Repository Structure](#repository-structure)
- [Research Documentation](#research-documentation)
- [Theory Testing](#theory-testing)
- [Research Methodology](#research-methodology)
- [Evidence and Sources](#evidence-and-sources)
- [Research Roadmap](#research-roadmap)
- [Project Status](#project-status)
- [Disclaimer](#disclaimer)
- [References](#references)

---

## Research Goals

The project aims to:

- Document publicly available information about VAC architecture and behavior.
- Compare reported behavior before and after significant updates.
- Record technical hypotheses and supporting evidence.
- Evaluate theories against available observations.
- Distinguish verified findings from assumptions.
- Maintain a structured research history that can be updated as new evidence becomes available.

Community reports and individual observations are not treated as definitive proof of VAC's internal implementation.

---

## Repository Structure

```text
vac-war/
├── README.md
├── research/
│   ├── pre-update/
│   │   └── first-theory/
│   │       └── theory-1.md
│   └── post-update/
│       └── first-theory/
│           └── theory-1.md
└── tests/
    └── theory-1-test.md
```

---

## Research Documentation

### Stage 1 — Pre-Update Analysis

Documentation of reported VAC architecture and behavior before the October 9, 2026 update.

- [Theory 1 — VAC Before the October 2026 Update](research/pre-update/first-theory/theory-1.md)

### Stage 1 — Post-Update Analysis

Investigation of reported behavioral changes following the October 9, 2026 update.

- [Theory 1 — VAC After the October 2026 Update](research/post-update/first-theory/theory-1.md)

Each theory should document its purpose, technical background, supporting sources, assumptions, limitations, and current confidence level.

---

## Theory Testing

Tests are maintained separately from theory documents so that hypotheses and their evaluations remain distinguishable.

### Test Registry

| Test | Purpose | Status |
|---|---|---|
| [Theory 1 Test](tests/theory-1-test.md) | Evaluate the first theory against available evidence and observations | Pending Validation |

### Test Status Definitions

- **Planned** — The test has been outlined but has not started.
- **In Progress** — Evidence is being collected or evaluated.
- **Pending Validation** — Initial analysis exists but requires further verification.
- **Supported** — Available evidence supports the hypothesis within the documented scope.
- **Inconclusive** — The evidence is insufficient to determine whether the hypothesis is correct.
- **Rejected** — Reliable evidence contradicts the hypothesis.

A test result must not be presented as proof of a general VAC detection or ban mechanism unless the evidence justifies that conclusion.

---

## Research Methodology

Research should follow a consistent process:

1. Define the question or hypothesis.
2. Record the relevant game version and update context.
3. Identify available sources and observations.
4. Separate direct evidence from interpretation.
5. Document alternative explanations and limitations.
6. Evaluate whether the evidence supports, contradicts, or fails to resolve the hypothesis.
7. Update the theory and test records when new information becomes available.

### Confidence Categories

- **Verified** — Supported by reliable sources or reproducible observations within a clearly defined scope.
- **Supported** — Consistent with available evidence, but not conclusively established.
- **Unconfirmed** — Plausible, but requires further investigation.
- **Speculative** — Based primarily on interpretation or incomplete information.
- **Rejected** — Contradicted by sufficiently reliable evidence.

These labels describe the strength of the available evidence, not an official VAC classification.

---

## Evidence and Sources

Whenever possible, research documents should include:

- Source links and publication dates.
- Relevant game versions and update dates.
- Screenshots, logs, or other supporting material when appropriate.
- An explanation of how each source supports the conclusion.
- Known limitations and possible alternative interpretations.

Community reports can help identify questions for investigation, but should not automatically be treated as verified technical findings.

Observations from one game version should not automatically be generalized to other versions.

---

## Research Roadmap

### Stage 1 — Initial Research

- [x] Establish the repository and documentation structure.
- [x] Separate pre-update and post-update theory documents.
- [ ] Review relevant public research and references.
- [ ] Document the evidence supporting the initial theories.
- [ ] Evaluate the first theory against available observations.
- [ ] Record limitations and unresolved questions.

### Future Research

- Expand the theory and test registry.
- Add further research topics when justified by new evidence.
- Improve source traceability and reproducibility.
- Revise earlier conclusions when stronger evidence becomes available.

The roadmap is provisional and may change as the project develops.

---

## Project Status

**Current Stage:** Stage 1 — Research and Analysis

**Overall Status:** Ongoing

**Primary Focus:** Documenting and evaluating theories concerning reported VAC behavior changes.

This repository is a work in progress. Conclusions may be revised, narrowed, or withdrawn when new evidence becomes available.

---

## Disclaimer

This repository is an independent educational and technical research project. It is not affiliated with or endorsed by Valve Corporation.

The documentation reflects publicly available information, observations, and hypotheses. It does not represent official documentation of VAC's internal implementation.

No specific detection mechanism, process indicator, or automatic ban condition should be considered confirmed without sufficient supporting evidence.

The presence of a process name, keyword, module, or individual technical indicator does not by itself establish that VAC will issue a ban.

All conclusions should be interpreted within the scope and limitations documented in the relevant research files.

---

## References

- [CS2 VAC Internals](https://github.com/Aspasia1337/cs2-vac-internals)
- [Official Counter-Strike 2 Updates](https://store.steampowered.com/news/app/730)
- [Public VAC Reverse-Engineering Research](https://github.com/search?q=VAC+anti-cheat+reverse+engineering&type=repositories)

Additional references should be added to the relevant research document when they support a specific claim.

---

*Independent research. Evidence over speculation.*
