console.log('foo');
require('./css/main.scss');
require('./css/unsemantic-grid-responsive.css')


// m.route(document.body, "/", {
//     "/": Layout(homepage),
//     '/app/new/webapp': Layout(NewWebapp),
//     '/app/:id': Layout(appview),
//     "/spore": Layout(spore),
//     '/spore/:id': Layout(spore)
// });
import React from 'react'
import req from 'superagent-bluebird-promise'
import { Router, Route, Link } from 'react-router'
import { history } from 'react-router/lib/HashHistory'
window.React = React

class LabeledInput extends React.Component {
  render() {
    return <div className='grid-100'>
      <div className='grid-100'><label>{this.props.label}</label></div>
      <input className='sp-input' type='text' onChange={::this.onChange} defaultValue={this.props.value}/>
    </div>
  }
  onChange(event) {
    if (this.props.onChange) {
      return this.props.onChange(event.target.value)
    }
  }
}

class WebappForm extends React.Component {
  constructor() {
    super()
    this.state = {
      Count: 2,
      ID: '',
      AttachedEnvs: [],
      ExtraEnv: {},
      Image: '',
      BalancedInternalTCPPort: 8000,
      Cpus: 2,
      Memory: 2048
    }
  }
  render() {
    console.log(this.state)
    var input = (prop, label) =>
      <LabeledInput label={label} onChange={this.inputChange(prop)} value={this.state[prop]}/>

    return <div>
      <h2 className='mono'>New Webapp</h2>
      {input('Count', 'Count')}
      {input('ID', 'ID')}
      {input('Image', 'Image')}
      {input('BalancedInternalTCPPort', 'Internal TCP Port')}
      {input('Cpus', 'CPUs')}
      {input('Memory', 'Memory')}
      <button className='sp-btn' onClick={::this.onSubmit}>Submit</button>
    </div>
  }
  inputChange(prop) {
    return (val) => {
      this.setState({[prop]: val})
    }
  }
  onSubmit() {
    return req.post('/api/v1/gen/webapp')
      .send({data: JSON.stringify(this.state)}).promise()
  }
}

class Sporedock extends React.Component {
  render() {
    return <div className='div.grid-container'>
      <div className='grid-100'>
        <h2>Sporedock</h2>
        <WebappForm/>
      </div>
    </div>
  }
}

React.render(
  (<Router history={history}>
    <Route path='/' component={Sporedock}/>
  </Router>), document.body)
