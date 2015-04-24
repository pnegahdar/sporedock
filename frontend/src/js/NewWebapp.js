var newWebapp = {};

newWebapp.controller = function () {
  newWebapp.vm.init();
};

newWebapp.vm = {};
newWebapp.vm.init = function () {
  this.form = {
    count: m.prop(0),
    id: m.prop(''),
    attachedEnvs: m.prop(''),
    extraEnvs: m.prop(''),
    image: m.prop(''),
    tcpPort: m.prop('80')
  };
};

newWebapp.vm.clickSubmit = function () {
  console.log(R.mapObj((fn) => fn(), this.form));
  // this.containers.push('B3');
};

newWebapp.view = function () {
  var vm = newWebapp.vm;

  var changer = (attr) => {
    return {onchange: m.withAttr('value', vm.form[attr]), value: vm.form[attr]()}
  };
  return (
    m('div.grid-100.grid-parent', [
      m('div.grid-50', [
        m('h2.mono', `New Webapp`)
      ]),
      m('div.grid-75.grid-parent', [
        m('div.grid-100', [m('label', 'Count')]),
        m('input.sp-input[type="text"]', changer('count')),
        m('div.grid-100', [m('label', 'ID')]),
        m('input.sp-input[type="text"]', changer('id')),

        m('div.grid-100', [m('label', 'Attached Envs')]),
        m('input.sp-input[type="text"]', changer('attachedEnvs')),
        m('div.grid-100', [m('label', 'Extra Env')]),
        m('textarea.sp-input.mono', changer('extraEnvs')),



        m('div.grid-100', [m('label', 'Image')]),
        m('input.sp-input[type="text"]', changer('image')),


        // m('div.grid-100', [m('label', 'Weight')]),
        // m('input.sp-input[type="text"]'),

        m('div.grid-100', [m('label', 'TCP Port')]),
        m('input.sp-input[type="text"]', changer('tcpPort')),
        m('div.grid-100', m('button.sp-btn', {onclick: vm.clickSubmit.bind(vm)}, 'Submit'))
      ])
    ])
  );
};

module.exports = newWebapp;
