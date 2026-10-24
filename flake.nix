# Nix flake: builds gitwhy from source.
#
# Use:  nix run github:dpologdvinity/gitwhy
#       nix profile install github:dpologdvinity/gitwhy
# Needs flakes enabled. The build needs no network beyond what Nix fetches itself.
{
  description = "gitwhy: do anything with git, see what it will do first, and undo it";

  # Unstable nixpkgs, because gitwhy needs a recent Go (go.mod sets the minimum).
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      # The same Linux and macOS targets the release builds.
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      # Runs f once per system with that system's package set.
      forAll = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAll (pkgs:
        let
          # A flake has no tag to read, so "gitwhy version" shows the commit hash, with "-dirty"
          # for uncommitted changes, or "dev" when there is no git info at all.
          version = self.shortRev or self.dirtyShortRev or "dev";
          gitwhy = pkgs.buildGoModule {
            pname = "gitwhy";
            inherit version;
            src = self;
            subPackages = [ "cmd/gitwhy" ];
            # Changes whenever go.mod or go.sum do: set vendorHash = pkgs.lib.fakeHash, build once and copy the hash from the error.
            vendorHash = "sha256-ohQZ2VWNBP/0eCQyNHHxcoSgu+GNAaCWBOYySVguK5A=";
            # Static binary, the same as the release build.
            env.CGO_ENABLED = 0;
            # Strip debug info and set the version "gitwhy version" prints.
            ldflags = [ "-s" "-w" "-X main.version=${version}" ];
            doCheck = false; # the end-to-end tests need real git repos and take minutes; CI runs them
            # Provides wrapProgram for postInstall.
            nativeBuildInputs = [ pkgs.makeWrapper ];
            # gitwhy runs the real git; use the user's if they have one, ours otherwise.
            # --suffix puts Nix's git at the end of PATH, so a git earlier in PATH wins.
            postInstall = ''
              wrapProgram $out/bin/gitwhy --suffix PATH : ${pkgs.lib.makeBinPath [ pkgs.git ]}
            '';
            meta = {
              description = "Do anything with git without looking it up, with previews and undo";
              homepage = "https://github.com/dpologdvinity/gitwhy";
              license = pkgs.lib.licenses.mit;
              # Lets "nix run" find the binary.
              mainProgram = "gitwhy";
            };
          };
        in
        # "default" makes "nix run" and "nix build" work without naming the package.
        { inherit gitwhy; default = gitwhy; });
    };
}
