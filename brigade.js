const {events, Job} = require("brigadier")

const moveAction = "action_move_card_from_list_to_list"

events.on("trello", (e, p) => {
  var hook = JSON.parse(e.payload)
  var d = hook.action.display

  if (d.translationKey = moveAction) {
    var e = d.entities
    console.log(`Card ${e.card.text} moved from ${e.listBefore.text} to ${e.listAfter.text}`)
  }

})
