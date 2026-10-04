# linter-autoconfigure-sdk — flake following the go-standard template
# (go-nix-helpers/templates/go-standard). Provides packages, apps,
# devShells (default/ci), and checks via the go-standard module.
#
# First landing: vendorHash is a placeholder — run
#   nix build && copy the `got:` hash
# then replace it below. Private deps are wired via `deps` so the FOD
# `go mod tidy` never touches the network and GOPRIVATE is auto-injected
# into devShells.
{
  description = "Shared foundation for linter auto-configuration tools — config round-trip, finding emission, and a provider spec for BuildFlow integration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "git+ssh://git@github.com/LarsArtmann/go-nix-helpers?ref=master";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    go-atomic-write = {
      url = "git+ssh://git@github.com/LarsArtmann/go-atomic-write?ref=refs/tags/v0.6.0";
      flake = false;
    };
    go-finding = {
      # toolsdk/v1.14.0 covers core v1.13.0 plus the toolsdk v1.14.0
      # required by go.mod (same single-ref pattern as erraudit).
      url = "git+ssh://git@github.com/LarsArtmann/go-finding?ref=refs/tags/toolsdk/v1.14.0";
      flake = false;
    };
    go-error-family = {
      # indirect in go.mod, but the FOD `go mod tidy` needs it resolvable.
      url = "git+ssh://git@github.com/LarsArtmann/go-error-family?ref=refs/tags/v0.11.0";
      flake = false;
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {
        pname = "linter-autoconfigure-sdk";
        vendorHash = "sha256-3w3bnc+bCHqsqHQuiR9cq5BQuCBPzb/gNyiyI30CxDk="; # nix build to compute
        description = "Shared foundation for linter auto-configuration tools";

        # go.mod floor is `go 1.27`; the module default (go_1_26 = 1.26.7)
        # fails under the devShell's GOTOOLCHAIN=local.
        goPkgAttr = "go_1_27";

        deps = {
          "github.com/larsartmann/go-atomic-write" = inputs.go-atomic-write;
          "github.com/larsartmann/go-finding" = inputs.go-finding;
          "github.com/larsartmann/go-error-family" = inputs.go-error-family;
        };
      };
    };
}
