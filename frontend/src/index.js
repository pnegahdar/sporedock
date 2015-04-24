console.log('foo');
require('./css/main.scss');

var spore = require('spore');
var homepage = require('home');
var appview = require('app');
var NewWebapp = require('NewWebapp');

var Layout = function(module) {
  return {
    controller: function() {
      return new Layout.controller(module)
    },
    view: Layout.view
  }
}
Layout.controller = function(module) {
  // Auth.loggedIn.then(null, function() {
  //   m.route("/login")
  // })
  this.content = module.view.bind(this, new module.controller)
}
Layout.view = function(ctrl) {
  return m('div.grid-container', [
    m('div.grid-100', [
      m('h2', {onclick: () => m.route('/')}, 'Sporedock')
    ]),
    ctrl.content()
  ]);
}


m.route(document.body, "/", {
    "/": Layout(homepage),
    '/app/new/webapp': Layout(NewWebapp),
    '/app/:id': Layout(appview),
    "/spore": Layout(spore),
    '/spore/:id': Layout(spore)
});

// m.render(document.body, layout.view());
