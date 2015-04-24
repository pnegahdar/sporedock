var spore = {};

spore.controller = function () {
  var id = m.route.param('id');
  spore.vm.init(id);
};

var getSpore = function () {
  var defer = m.deferred();
  setTimeout(() => defer.resolve(['C1', 'C2', 'C3', 'C4']), 300);
  return defer.promise;
}

spore.vm = {};
spore.vm.init = function (id) {
  this.id = id;
  this.containers = m.prop([]);
  getSpore().then(this.containers).then(m.redraw);
};

spore.vm.clickContainer = function (container) {
  console.log(container);
  // this.containers.push('B3');
};

spore.vm.clickSpore = function (container) {
  console.log(container);
  m.route(`/spore/${container.id}`);
};

spore.view = function () {
  var vm = spore.vm;
  return (
    m('div.grid-100.grid-parent.sp-spores', [
      m('div.grid-50', [
        m('h2.sp-app-name.mono', `Spore ${vm.id}`)
      ]),
      m('div.grid-50', [
      ]),
      m('div.grid-100', [m('h3', 'Spores')]),
      m('div.grid-100', [
        vm.containers().map(container =>
          m('div.sp-docker', {onclick: vm.clickContainer.bind(vm, container)}, container))
      ])
    ])
  );
};

module.exports = spore;
