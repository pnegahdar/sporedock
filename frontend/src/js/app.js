var appview = {};

appview.controller = function () {
  appview.vm.init(m.route.param('id'));
};

var getApp = function (id) {
  var defer = m.deferred();
  setTimeout(() => defer.resolve({
    id: id + Math.random(),
    count: 5,
    attachedEnvs: ['prod'],
    extraEnv: {GIT_REV: '3ac3df'},
    tags: {created_at: '2012-03-05T03:02:01'},
    image: 'gallery.aws.kensho.com:5000/kensho/webfe:latest',
    weight: 2.0,
    balancedTCPPort: 80,
    status: 'pulling',
    spores: [
      {id: 'a3f', name: 'Spore 1', containers: ['C1', 'C2', 'C3', 'C4']},
      {id: 'b4c', name: 'Spore 2', containers: ['C1', 'C2', 'C3']},
      {id: 'c2c', name: 'Spore 3', containers: ['C1', 'C2']}
    ]
  }), 200);
  return defer.promise;
}

appview.vm = {};
appview.vm.init = function (id) {
  this.id = id;
  this.app = m.prop({id: id, spores: []});
  setInterval(() => getApp(id).then(this.app).then(m.redraw), 2000);
};

appview.vm.clickContainer = function (container) {
  console.log(container);
};

appview.vm.clickSpore = function (container) {
  console.log(container);
  m.route(`/spore/${container.id}`);
};

appview.view = function () {
  var vm = appview.vm;
  var app = vm.app();
  return (
    m('div.grid-100.grid-parent.sp-spores', [
      m('div.grid-50', [
        m('h2.sp-app-name.mono', app.id)
      ]),
      m('div.grid-50', [
      ]),
      m('div.grid-100', [m('h3', 'Spores')]),
      m('div.grid-100', [
        app.spores.map(spore =>
          m('div.sp-spore', [
            m('h4', {onclick: vm.clickSpore.bind(vm, spore)}, spore.name),
            spore.containers.map(container =>
              m('div.sp-docker', {onclick: vm.clickContainer.bind(vm, container)}, container))
          ]))
      ])
    ])
  );
};

module.exports = appview;
