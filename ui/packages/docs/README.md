# Website

This website is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.

### Character script reference data

The script reference JSON files are `src/components/{Names,Actions,Params,Fields,Issues}/character.dm.json`.
Their source is `internal/characters/**/config.yml`. The manifests in
`character_imports` and `pipeline/community/data` track imported game data and
review status, not script syntax.

From the repository root, regenerate just these reference files with Python 3
and PyYAML (no datamine download required):

```sh
python3 scripts/generate_character_docs.py
python3 scripts/generate_character_docs.py --check
```

`task character-docs` runs the same generator. It uses the existing pipeline JSON
schemas and preserves the distinction between supported syntax and verified
mechanics. When adding a character or action parameter, update its `config.yml`
before regenerating. The full `task pipeline` remains the generator for talent
tables and the other game data.

### Installation

```
$ yarn
```

### Local Development

```
$ yarn start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

### Build

```
$ yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

### Deployment

Using SSH:

```
$ USE_SSH=true yarn deploy
```

Not using SSH:

```
$ GIT_USER=<Your GitHub username> yarn deploy
```

If you are using GitHub pages for hosting, this command is a convenient way to build the website and push to the `gh-pages` branch.
