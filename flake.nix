{
  description = "develop environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { nixpkgs, ... }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
    in
    {
      devShells = nixpkgs.lib.genAttrs systems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          go = pkgs.go_1_27;
        in
        {
          default = pkgs.mkShell {
            packages = [ go ];

            # Print only `go: <version>` so `nix develop` stays terse.
            # `go env GOVERSION` reports the version of the installed binary
            # (for example go1.27.3); the `${v#go}` expansion strips "go".
            shellHook = ''
              v="$(go env GOVERSION)"
              echo "go: ''${v#go}"
            '';
          };
        }
      );
    };
}
