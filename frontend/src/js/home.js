var homepage = {};

homepage.controller = function () {
  homepage.vm.init();
};

var getApps = function () {
  var defer = m.deferred();
  setTimeout(() => defer.resolve([
    {id: 'kensho-eventq-webfe', name: 'App 1', containers: ['C1', 'C2', 'C3', 'C4']},
    {id: 'b4c', name: 'App 2', containers: ['C1', 'C2', 'C3']},
    {id: 'c2c', name: 'App 3', containers: ['C1', 'C2']}
  ]), 200);
  return defer.promise;
}

homepage.vm = {};
homepage.vm.init = function () {
  this.apps = m.prop([]);
  getApps().then(this.apps).then(m.redraw);
};

homepage.vm.clickContainer = function (container) {
  console.log(container);
};

homepage.vm.clickApp = function (container) {
  console.log(container);
  m.route(`/app/${container.id}`);
};

homepage.view = function () {
  var vm = homepage.vm;
  return (
    m('div.grid-100.grid-parent.sp-spores', [
      m('div.grid-100', [m('h3', 'Apps')]),
      m('div.grid-100', [
        vm.apps().map(app =>
          m('div.sp-app', [
            m('h4', {onclick: vm.clickApp.bind(vm, app)}, app.name),
          ]))
      ])
    ])
  );
};

module.exports = homepage;
