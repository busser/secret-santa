# Secret Santa

Secret Santa is a game where each member of the family is assigned another
member of the family to give a gift to for Christmas. Each pairing is known only
to the person making the gift. Nobody knows who is getting them a gift, hence
the name _Secret Santa_.

## Why does this tool exist?

When organising a Secret Santa with my family, we realised we wanted to respect
certain conditions that existing Secret Santa services could not provide. The
constraints are:

1. The chain of Santas makes a single large loop.
2. Nobody gets assigned someone in their immediate family.
   - Our extended family is composed of 4 "sub-families".

## Usage

To build the tool, you need to have [Go](https://golang.org/doc/install)
installed. Then, run this command:

```bash
make build
```

Run the tool by passing it a configuration file:

```bash
bin/secret-santa --config=secret-santa.yml
```

The configuration file should look something like this:

```yaml
families:
  - name: Li
    members:
      - Ang
      - Eugenia
      - Jackie
  - name: Stone
    members:
      - John
      - Mia
  - name: Virtue
    members:
      - Desiree
      - Trevor
```
