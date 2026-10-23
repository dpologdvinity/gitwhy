{
  description = "gitwhy: do anything with git, see what it will do first, and undo it";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAll = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAll (pkgs:
        let
          version = self.shortRev or self.dirtyShortRev or "dev";
          gitwhy = pkgs.buildGoModule {
            pname = "gitwhy";
            inherit version;
            src = self;
            subPackages = [ "cmd/gitwhy" ];
            # Changes whenever go.mod or go.sum do: set vendorHash = pkgs.lib.fakeHash, build once and copy the hash from the error.
            vendorHash = "sha256-ohQZ2VWNBP/0eCQyNHHxcoSgu+GNAaCWBOYySVguK5A=";
            env.CGO_ENABLED = 0;
            ldflags = [ "-s" "-w" "-X main.version=${version}" ];
            doCheck = false; # the end-to-end tests need real git repos and take minutes; CI runs them
            nativeBuildInputs = [ pkgs.makeWrapper ];
            # gitwhy runs the real git; use the user's if they have one, ours otherwise.
            postInstall = ''
              wrapProgram $out/bin/gitwhy --suffix PATH : ${pkgs.lib.makeBinPath [ pkgs.git ]}
            '';
            meta = {
              description = "Do anything with git without looking it up, with previews and undo";
              homepage = "https://github.com/dpologdvinity/gitwhy";
              license = pkgs.lib.licenses.mit;
              mainProgram = "gitwhy";
            };
          };
        in
        { inherit gitwhy; default = gitwhy; });
    };
}
