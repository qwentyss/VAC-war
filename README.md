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
| [Post-Update Theory 1](research/post-update/first-theory/theory-1.md) | Hypothesis concerning possible process-name inspection | Unverified |
| [Exploratory Keywords](research/post-update/first-theory/keywords.md) | Unverified terms discussed in the hypothesis | Unverified |

## Test Registry

| Test | Related Research | Status |
|---|---|---|
| [Post-Update Test 1](tests/post-update/test-1/README.md) | [Post-Update Theory 1](research/post-update/first-theory/theory-1.md) | In progress — preliminary observation recorded |

## Current Project Status

**Current Stage:** Stage 1 — Research and Analysis

**Overall Status:** Stoped

**Phase:** Post-Update Observation and Preliminary Evaluation

**Test window:** October 16–25, 2026

**Planned sample:** 15 matches

The first post-update test began on October 16, 2026. An initial observation has been recorded. The match was interrupted, and a ban was reported after the session.

The available observation does not establish the cause of the ban or the mechanism responsible for it. It does not support the expectation that renaming the software would prevent enforcement under the conditions observed.

The hypothesis remains under evaluation. Any final assessment must account for the available evidence, incomplete sessions, and alternative explanations.

### Progress Checklist

- [x] Create the research and test documentation
- [x] Record initial system and game information
- [x] Complete and document environment verification
- [x] Complete the planned observations, or document why testing stopped
- [x] Review and consolidate the observation log
- [x] Evaluate the findings and limitations
- [x] Publish the final assessment

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

A single observation may contradict a hypothesis without conclusively disproving every version of it. Conclusions should reflect the strength and limitations of the available evidence.

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

Research conclusions are limited by the available evidence, recorded conditions, and observation period. Testing must comply with applicable rules, terms of service, and authorization requirements.
