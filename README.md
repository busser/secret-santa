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

The tool works in two steps. First, draw the Santa chain:

```bash
bin/secret-santa draw
```

This step picks a giftee and a message from `messages.yml` for each Santa. It
saves the texts to `draw.json`. It does not show who gives a gift to whom, so
the organizer can also take part. It refuses to replace an existing
`draw.json`. To draw again, delete the file.

Then, send each Santa their text by SMS through OVHcloud:

```bash
direnv exec . bin/secret-santa send
```

This step reads the OVH credentials from the environment (see `.envrc`). It
records each text that it sends in `draw.json`. If a text fails, run the
command again: it sends only the texts that were not sent. Phone numbers come
from the configuration file, so you can correct a number after the draw.

The configuration file should look something like this:

```yaml
families:
  - name: Li
    members:
      - name: Ang
        phone: +1234567890
      - name: Eugenia
        phone: +2345678901
      - name: Jackie
        phone: +3456789012
  - name: Stone
    members:
      - name: John
        phone: +4567890123
      - name: Mia
        phone: +5678901234
  - name: Virtue
    members:
      - name: Desiree
        phone: +6789012345
      - name: Trevor
        phone: +7890123456
```
