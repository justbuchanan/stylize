{
  description = "Stylize - code formatting tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "stylize";
          version = "0.1.0";

          src = ./.;

          vendorHash = "sha256-rBYSRoGhq5RqP4f50Vw78JYPO/6YKQo5q6ckdmPPlJc=";
          goSum = ./go.sum;

          # Don't run tests as part of the build
          doCheck = false;

          meta = with pkgs.lib; {
            description = "Quickly reformats or checkstyles an entire repository of code";
            homepage = "https://github.com/justbuchanan/stylize";
            license = licenses.asl20;
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            python3Packages.yapf
            python3Packages.black
            uncrustify
            buildifier
            prettier
          ];

          shellHook = ''
            echo "Development environment for stylize"
            echo "Go version: $(go version)"
          '';
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/stylize";
        };
      }
    );
}
