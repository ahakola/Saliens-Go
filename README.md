# Saliens-Go
## 2018 Summer Saliens - Bosswatch
Detecting bosses on active planets during Steam's 2018 "Intergalactic Summer Sale".

This app polls Steam's API for information about state of the "Summer Saliens"-game and parses json-responses to detect if these high experience yielding boss monsters have spawned or died and alert user. Written in Go.

### Example output:
```
------------------------------------------------------------
2018 Summer Saliens - Bosswatch
                                      I'm hunting wabbits...
------------------------------------------------------------

[02.07.2018 12:16:50] Script started.

------------------------------------------------------------
[02.07.2018 13:37:09] >>> BOSS DETECTED <<<

- Zone 23 on planet 'Reboots Planet II' (Id: 526)
- First boss detected 1h20m19s after the start

------------------------------------------------------------
[02.07.2018 13:46:26] <<< BOSS IS GONE >>>

- Encounter lasted ~9m17s
- Estimate for maximum EXP gain is 278500 + bonuses

------------------------------------------------------------
[02.07.2018 16:30:23] >>> BOSS DETECTED <<<

- Zone 19 on planet 'Take it Easy Planet II' (Id: 41)
- 2 bosses detected in 4h13m33s
- Last boss killed 2h43m58s ago

------------------------------------------------------------
[02.07.2018 16:37:40] <<< BOSS IS GONE >>>

- Encounter lasted ~7m17s
- Estimate for maximum EXP gain is 218500 + bonuses
```

### If you use API, why repository includes JSON-files?

JSON-files are provided in case someone wants to play with the code after the Steam's 2018 "Intergalactic Summer Sale"-event has ended and you can't get proper game state data from API anymore.