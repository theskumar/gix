# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Implemented a new AI-driven command for generating changelogs.
- Added support for the Google Gemini AI provider.
- Enabled interactive review and editing of suggested commits.

### Changed

- Improved AI efficiency by compressing staged diff input to save tokens.
- Enhanced the consistency and quality of AI-generated output by providing detailed Conventional Commits specifications to the model.

### Fixed

- Fixed an issue where `git diff` hunks were not reliably parsed during patching and splitting operations.
