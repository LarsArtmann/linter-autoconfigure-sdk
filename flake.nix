# linter-autoconfigure-sdk — flake following the go-standard template
# (go-nix-helpers/templates/go-standard). Provides packages, apps,
# devShells (default/ci), and checks via the go-standard module.
#
# Deps are wired via `deps` even though the repos are public: the module
# always marks `larsartmann/*` as private, so a deps-less FOD bypasses the
# proxy for direct VCS (blocked in the sandbox). With `deps`, the FOD
# resolves everything from the local prepared source — no network needed,
# and `github:` inputs keep it CI-safe without SSH keys.
{
  description = "Shared foundation for linter auto-configuration tools — config round-trip, finding emission, and a provider spec for BuildFlow integration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "github:LarsArtmann/go-nix-helpers";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    go-atomic-write = {
      url = "github:LarsArtmann/go-atomic-write?ref=refs/tags/v0.6.0";
      flake = false;
    };
    go-finding = {
      # toolsdk/v1.14.0 covers core v1.13.0 plus the toolsdk v1.14.0
      # required by go.mod (same single-ref pattern as erraudit).
      url = "github:LarsArtmann/go-finding?ref=refs/tags/toolsdk/v1.14.0";
      flake = false;
    };
    go-error-family = {
      # indirect in go.mod, but the FOD `go mod tidy` needs it resolvable.
      url = "github:LarsArtmann/go-error-family?ref=refs/tags/v0.11.0";
      flake = false;
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {
        pname = "linter-autoconfigure-sdk";
        vendorHash = "sha256-3w3bnc+bCHqsqHQuiR9cq5BQuCBPzb/gNyiyI30CxDk=";
        description = "Shared foundation for linter auto-configuration tools";

        # go.mod floor is `go 1.27`; the module default (go_1_26 = 1.26.7)
        # fails under the devShell's GOTOOLCHAIN=local.
        goPkgAttr = "go_1_27";

        deps = {
          "github.com/larsartmann/go-atomic-write" = inputs.go-atomic-write;
          "github.com/larsartmann/go-finding" = inputs.go-finding;
          "github.com/larsartmann/go-error-family" = inputs.go-error-family;
        };

        # TestREADMESnippetsCompile compiles README snippets in a temp module
        # that cannot resolve private deps under the sandbox's GOPROXY=off
        # (locally they resolve via direct VCS fetch). Same class as pdd's
        # check disablement: tests run in the devShell, not the hermetic build.
        enableCheck = false;
      };
    };
}
