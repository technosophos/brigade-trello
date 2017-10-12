const {events, Job} = require("brigadier")

events.on("trello", (e, p) => {
  console.log(e.payload)
})
