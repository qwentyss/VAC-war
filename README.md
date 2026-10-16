# VAC-war

Independent research repository documenting hypotheses, observations, and testing related to reported changes in Counter-Strike 2's Valve Anti-Cheat (VAC) behavior.

## Project Goals

- Document hypotheses concerning reported VAC behavior changes.
- Separate verified information from speculation.
- Maintain structured research notes and observation records.
- Evaluate hypotheses against evidence rather than assumptions.
- Record limitations and contradictory findings.

## Repository Structure

```text
VAC-war/
├── README.md
├── research/
│   ├── pre-update/
│   │   └── first-theory/
│   │       └── theory-1.md
│   └── post-update/
│       └── first-theory/
│           ├── theory-1.md
│           ├── theory-2.md
│           └── keywords.md
└── tests/
    └── post-update/
        └── test-1/
            ├── README.md
            ├── test-plan.md
            ├── test-environment.md
            └── results.md
```

*This tree should match the files actually present in the repository.*

## Research Registry

| Research | Description | Status |
|---|---|---|
| [Pre-Update Theory 1](research/pre-update/first-theory/theory-1.md) | Preliminary overview of possible VAC detection mechanisms | Preliminary |
| [Post-Update Theory 1](research/post-update/first-theory/theory-1.md) | Hypothesis concerning possible process-name inspection | Not supported by the available test evidence |
| [Post-Update Theory 2](research/post-update/first-theory/theory-2.md) | Conceptual hypothesis about context-aware ESP and in-game perception | Proposed — unverified |
| [Exploratory Keywords](research/post-update/first-theory/keywords.md) | Unverified terms examined in connection with Theory 1 | Preliminary findings documented |

## Test Registry

| Test | Related Research | Status |
|---|---|---|
| [Post-Update Test 1](tests/post-update/test-1/README.md) | [Post-Update Theory 1](research/post-update/first-theory/theory-1.md) | Stopped — preliminary findings recorded |

Theory 2 has not been established by the observations from Test 1. It should be evaluated separately if additional research is conducted.

## Current Project Status

**Current Stage:** Stage 1 — Research and Analysis

**Overall Status:** Ongoing

**Phase:** Hypothesis Documentation and Evidence Review

### Theory 1 — Preliminary Findings

The first post-update test examined whether renaming software and checking proposed keyword variations would prevent a ban.

A ban was reported after an interrupted match on October 16, 2026. The planned sample of 15 matches was not completed.

The observation does not support the expectation that renaming alone would prevent enforcement under the recorded conditions. The exact cause of the reported ban has not been independently established.

**Current assessment:** Theory 1 is not supported by the available test evidence. The finding is preliminary and does not establish VAC's internal detection mechanism.

### Theory 2 — Proposed Hypothesis

Theory 2 considers a conceptual form of context-aware ESP whose displayed information is limited according to in-game visibility, visual impairment, and naturally available information.

This theory is currently unverified. No conclusion has been established about its technical feasibility or relationship to anti-cheat enforcement.

### Project Checklist

- [x] Document the initial research hypotheses
- [x] Record the available information from the first test
- [x] Document the preliminary findings from Test 1
- [x] Create a separate document for Theory 2
- [ ] Review sources supporting or contradicting Theory 2
- [ ] Document evidence and limitations for each theory
- [ ] Update conclusions when sufficient evidence becomes available

Checklist items should be marked complete only when the corresponding work has actually been completed.

## Research Methodology

Research documents should distinguish between the following categories:

- **Verified information:** Information supported by reliable sources or reproducible evidence.
- **Hypothesis:** A proposed explanation that has not been established.
- **Observation:** An event or result directly recorded during research.
- **Interpretation:** An explanation proposed for an observation.
- **Limitation:** A factor that restricts the conclusions that can be drawn.

A lack of visible detection or enforcement does not establish that VAC did not detect something internally.

A ban or other enforcement event does not, by itself, identify the mechanism responsible.

A single observation may contradict a hypothesis without conclusively disproving every possible version of it. Conclusions should reflect the strength and limitations of the available evidence.

## Evidence Standards

- Cite sources for technical claims wherever possible.
- Record relevant dates, software versions, and test conditions.
- Distinguish correlation from causation.
- Document contradictory observations and interrupted sessions.
- Keep observations separate from interpretations.
- Do not treat community reports as independently verified facts.
- Do not claim that a hypothesis is confirmed or conclusively disproved without sufficient evidence.
- Preserve the distinction between a reported event and an independently verified explanation.

## Sources

- [Counter-Strike 2 Official Updates](https://store.steampowered.com/news/app/730)
- [CS2 VAC Internals — Community Research](https://github.com/Aspasia1337/cs2-vac-internals)
- [GitHub Search: VAC Anti-Cheat Research](https://github.com/search?q=VAC+anti-cheat+reverse+engineering&type=repositories)

These sources provide research context and do not necessarily validate every claim in this repository. Individual claims should be supported by relevant evidence.

## Disclaimer

This repository is an independent research project and is not affiliated with or endorsed by Valve.

The hypotheses and observations documented here are not official statements about VAC's internal implementation. A claim should not be considered established solely because it appears in this repository.

Research conclusions are limited by the available evidence, recorded conditions, and observation period. Research must comply with applicable rules, terms of service, and authorization requirements.
