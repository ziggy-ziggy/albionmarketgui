Albion Market GUI is a WIP project, intended to display market data in an easy format for market flipping between towns.

This project needs a lot of improvements, and while it was fun to learn Go, I need to shelf this project for a while. Uploading this to Github so I can come back to this later, or maybe someone else can use it.

![Imgur Image](https://i.imgur.com/x3piAcK.png)
Some notable features:
Display useful market data, such as profit margins or profit in a single trip between the market with the lowest and highest cost for that item. Those costs are highlighted.
Select which items to display by their categories, which can be configured in the itemsData.yaml file by changing the item's Toggle value
Old market data is greyed out (In the screenshot I haven't visited Martlock for 48+ hours)
Option to ignore Caerleon so you can purely trade in safe zones.

---

I don't recommend anyone use this tool in its current state, but if you do there's a couple of hard coded configurations you'll have to do.

To run this, you'll need to make a mySQL database. Then put that database's username and password in the "Setup database connection" part of the albiondata-client.go file. The table should be made automatically when starting the client.

While the application is running, you should see the GUI application from the screenshot. If you visit the markets and check the orders (I recommend using the "Create Buy Order" tab), and then click confirm on the GUI. If the table is connected properly, the data should show up for the items you checked. Note that there's currently a bug and you have to scroll the table before anything shows up.

Next, to configure the colors for the profit margin/profit per trip columns, adjust the marginConst and tripConst constants in the albiondata-client.go file.

---

TODO List:
1. Right now the client is just my GUI application run in parallel with the Albiondata-client so I could just use local tables. Should just make it use the data uploaded to the server and isolate the GUI application
2. The methods for distinguishing items between its separate tiers probably needs more improvement. I haven't even tested if it works for all of the items.
3. Updating to use the market history could give some useful insights.
4. A lot of values are hardcoded (color values, mySQL database connection), these should have GUI elements.
5. More error handling
6. There's probably tons of performance improvements to be made.
7. A lot of GUI improvements to be made. The GUI framework, Fyne, feels pretty early in development and pretty limiting. Maybe it'll give some better tools later.

And whatever else I can't think of off the top of my head

# License
This project, and all contributed code, are licensed under the MIT
License. A copy of the MIT License may be found in the repository.
