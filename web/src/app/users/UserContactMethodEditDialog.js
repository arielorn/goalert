import React from 'react'
import p from 'prop-types'
import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import Query from '../util/Query'
import { omit } from 'lodash-es'

const query = gql`
  query($id: ID!) {
    userContactMethod(id: $id) {
      id
      name
      type
      value
    }
  }
`

const mutation = gql`
  mutation($input: UpdateUserContactMethodInput!) {
    updateUserContactMethod(input: $input)
  }
`

export default class UserContactMethodEditDialog extends React.PureComponent {
  static propTypes = {
    cmID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: null,
    errors: [],
    edit: true,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.cmID }}
        render={({ data }) => this.renderMutation(data.userContactMethod)}
        noPoll
      />
    )
  }

  renderMutation({ name, type, value }) {
    return (
      <Mutation
        client={graphql2Client}
        mutation={mutation}
        awaitRefetchQueries
        refetchQueries={['cmList']}
        onCompleted={this.props.onClose}
      >
        {(commit, status) =>
          this.renderDialog(commit, status, { name, type, value })
        }
      </Mutation>
    )
  }

  renderDialog(commit, status, defaultValue) {
    const { loading, error } = status

    const fieldErrs = fieldErrors(error)

    return (
      <FormDialog
        title='Edit Contact Method'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              // removing field 'type' from value for mutation
              input: { ...omit(this.state.value, 'type'), id: this.props.cmID },
            },
          })
        }}
        form={
          <UserContactMethodForm
            errors={fieldErrs}
            disabled={loading}
            edit={this.state.edit}
            value={this.state.value || defaultValue}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }
}
