/* eslint-env browser */
const socket = new WebSocket(location.origin.replace('http:', 'ws:') + '/fu')

const status = document.getElementById('status')
const output = document.getElementById('output')
const form = document.querySelector('form')
const buttons = form.querySelectorAll('button')

socket.addEventListener('open', function () {
  status.innerHTML = 'connected'
})

socket.addEventListener('message', function (e) {
  output.value = e.data + '\n' + output.value
})

socket.addEventListener('close', function () {
  status.innerHTML = 'connection closed'
})

buttons.forEach(function (btn) {
  btn.addEventListener('click', function (e) {
    e.preventDefault()
    socket.send(e.target.innerText)
  })
})
