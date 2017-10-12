const {events, Job} = require("brigadier")

const moveAction = "action_move_card_from_list_to_list"

events.on("trello", (e, p) => {
  var hook = JSON.parse(e.payload)
  var d = hook.action.display

  if (d.translationKey = moveAction) {
    console.log(`Card ${d.card.text} moved from ${d.listBefore.text} to ${d.listAfter.text}`)
  }

})
