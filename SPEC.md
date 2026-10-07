# Take-home Project: Multiplayer Conway's Game of Life

Conway's Game of Life is a famous simulation that demonstrates a cellular automaton. It is modeled as a grid with 4 simple rules.

## Rules

1. Any live cell with fewer than two live neighbors dies, as if caused by under-population.
2. Any live cell with two or three live neighbors lives on to the next generation.
3. Any live cell with more than three live neighbors dies, as if by overcrowding.
4. Any dead cell with exactly three live neighbors becomes a live cell, as if by reproduction.

Create a web app version of Game of Life, with the following functions. At minimum, you must support the latest Google Chrome desktop browser, although you may be penalized if your game doesn't work in other modern browsers. (For example, we grade using Google Chrome, but clients might open your work in Firefox or Safari.)

## Multiplayer

1. Implement the Game of Life browser frontend. You can use any representation such as `<canvas>`, simple DOM manipulation, or even `<table>` cells. The game should tick automatically at a predefined interval, at say, 1 step per second.
2. The browser connects to a server, which allows multiple browsers to share the same, synchronized world view. Unless otherwise specified, the server may be written in Python, Node.js, or any other technology which can be run on Linux.
3. Each client is assigned a random color on initialization. From the browser, clicking on any grid will create a live cell on that grid with the client's color. This change should be synchronized across all connected clients. (You can use any mechanism to achieve this, such as polling, SSE, or WebSocket.)
4. When a dead cell revives by rule 4, "Any dead cell with exactly three live neighbors becomes a live cell, as if by reproduction.", it will be given a color that is the average of its neighbors (that revive it).
5. To make the evolution more interesting, include a toolbar that places some predefined patterns at random places with the player's color, such as those found at <https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life#Examples_of_patterns> (not necessary to implement all, just 3–4 is fine).

## Project submission

We expect you to keep track of your progress as you work on the project using Git, and share the repository with us when you finish the project.

## Skills to be graded

1. Algorithmic Programming
2. API and Service Design
3. Testing, CI/CD, and Site Reliability
4. UX Design and Prototyping
5. Web Frontend
6. Communication

## README

Write your README as if it was for a production service and the only document available for other developers. Include the following items:

- Description of the problem and solution.
- How to test, build, deploy, and use your solution.
- Reasoning behind your technical choices, including architectural.
- Trade-offs you might have made, anything you left out, or what you might do differently if you were to spend additional time on the project.
- Link to other projects or code you're particularly proud of.

The easier it is to understand your code the better grade it would have.

Pay specific attention to the following:

1. Correctness / Robustness. For example, does it behave correctly if a client's connection is unstable for a few seconds? Can we easily build and deploy?
2. Performance / Scalability, both on the client side and the server side.
3. Code style / terseness / proper use of latest technologies.
4. UI aesthetics.
5. Communication skills and documentation.
6. Your code should preferably have a Dockerfile for ease of deployment.
7. If you run out of time, document what features are missing, and how you would approach them if you have more time. Please spend no more than a few hours.
8. Be ready to walk us through your solution and defend any part of it during your follow-up presentation.

## AI usage

Using AI will not count against you, but you will need to document its use, just like your usage of any other tools. Include in the README:

- Which tools (harnesses, models, skills) you used.
- Your overall workflow: what you wrote, what you decided, what you delegated.
- The key prompts you entered. (A session transcript would be nice.)

We will evaluate how well you direct it and judge its output.

## FAQ

**The description says "Each client is assigned a random color on initialization". If the client disconnects and reconnects some time in the future, do they need to keep the same color?**

This is unspecified. Use your best judgment. When we grade your solution, we will also look at your engineering choices and trade-offs.

**What do you mean by the average of colors?**

The mathematical average is sufficient, although feel free to use other definitions of color averaging. Remember to document your choices.

**How large is the playground? May I just put the width and height in the config file? Or do you need it to be full-size of the browser's screen, or even resizable?**

We have received submissions with board sizes ranging from a 10×10 board to millions of cells. Please use your best judgement. Hint: many employers value readability and maintainability over fancy features. The choice of algorithms that are demonstrably scalable is also a big plus.

**What's the scale of the game? How many clients do you expect at the same time? Should I limit the client's count?**

We expect something that can be achieved within a few hours. We don't expect a reasonably sized MMO.

**What's the initial state on the playground? Do we need to randomly initialize some cells? Or just leave it blank?**

It is acceptable to have a blank world as the initial state.

**When the predefined pattern put by a user covers an alive cell, I guess we just overwrite the cell's color by the pattern's color, am I correct?**

This is an acceptable strategy. You may also want to put this and any other assumptions you have made that were not in the specification into the documentation.

**Some parts of the spec are incomplete?**

Hint: in real-world situations, you will rarely encounter a 100% complete spec. In these cases, you'll need to use your communications skills to either make reasonable assumptions or clarify with the relevant parties.

**This is a game. Tell me the most attractive point of the game/the product to ensure I got the point.**

The point is for you to showcase your skills and abilities while also having a bit of fun!

Copyright 2016–2021 Terminal 1 Limited.
Copyright 2026 hermeneutic Research, LLC.
