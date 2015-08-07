require('./css/main.scss');
require('./css/unsemantic-grid-responsive.css')

import React from 'react'
import req from 'superagent-bluebird-promise'
import { Router, Route, Link, Navigation } from 'react-router'
import { history } from 'react-router/lib/HashHistory'
import { combineReducers, createStore, applyMiddleware } from 'redux'
import { Provider, connect } from 'react-redux'
import R from 'ramda'
import thunk from 'redux-thunk'
import * as reducers from './reducers'
import * as actions from './actions/webapp'
window.React = React

let createStoreWithMiddleware = applyMiddleware(thunk)(createStore)
let webapps = combineReducers(reducers)
let store = createStoreWithMiddleware(webapps)

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
      ID: 'parhambox',
      AttachedEnvs: [],
      ExtraEnv: {},
      Image: '',
      BalancedInternalTCPPort: 8000,
      Cpus: 2,
      Memory: 2048
    }
  }
  render() {
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
      <button className='sp-btn' onClick={R.partial(this.props.onSubmit, this.state)}>Submit</button>
    </div>
  }
  inputChange(prop) {
    return (val) => {
      this.setState({[prop]: val})
    }
  }
}

var Webapp = connect(state => state.webapp)(React.createClass({
  mixins: [Navigation],
  render() {
    return <div>
      <h2 className='mono'>Webapp {this.props.params.id}</h2>
      {JSON.stringify(this.props)}
      <div>
        <button className='sp-btn' onClick={this.clickDelete}>Delete</button>
      </div>
    </div>
  },
  clickDelete() {
    this.props.dispatch(actions.deleteWebapp(this.props.params.id))
      .then(() => this.transitionTo('/'))
  }
}))

var WebappList = React.createClass({
  getInitialState() {
    return {apps: []}
  },
  render() {
    return <div>
      <h2 className='mono'>All Webapps</h2>
      <div>
        {this.props.apps.map(app => <div><Link to={`/webapp/${app.ID}`}>ID: {app.ID}</Link></div>)}
      </div>
    </div>
  }
})

@connect(state => state.webapp)
class Sporedock extends React.Component {
  constructor() {
    super()
  }
  componentDidMount() {
    this.props.dispatch(actions.getWebappList())
  }
  render() {
    console.log('render', this.props)
    return <div className='div.grid-container'>
      <div className='grid-100'>
        <h2><Link to={'/'}>Sporedock</Link></h2>
        <WebappList apps={this.props.apps}/>
        {this.props.children || <WebappForm onSubmit={::this.onSubmit}/>}
      </div>
    </div>
  }
  onSubmit(data) {
    this.props.dispatch(actions.newWebapp(data))
      }
}

function routes() {
  return <Router history={history}>
    <Route path='/' component={Sporedock}>
      <Route path='webapp/:id' component={Webapp}/>
    </Route>
  </Router>
}

React.render(
  (<Provider store={store}>{routes}</Provider>), document.body)
