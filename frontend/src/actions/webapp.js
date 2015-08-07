import req from 'superagent-bluebird-promise'
import R from 'ramda'

function getAppList() {
  return req.get('/api/v1/gen/webapp')
    .then(R.path(['body', 'data']))
}

export function newWebapp(params) {
  console.log('newWebapp', params)
  return dispatch => {
    return req.post('/api/v1/gen/webapp')
      .send({data: JSON.stringify(params)})
      .then(() => getWebappList()(dispatch))
  }
}

export function deleteWebapp(id) {
  return dispatch => {
    return req.del(`/api/v1/gen/webapp/${id}`)
      .then(() => getWebappList()(dispatch))
  }
}



export function getWebappList() {
  return dispatch => {
    return getAppList().then(data => dispatch(updateWebappList(data)))
  }
}

export function updateWebappList(params) {
  return {
    type: 'APP_LIST',
    params
  }
}
