import React from 'react'
import p from 'prop-types'
import Query from '../util/Query'
import gql from 'graphql-tag'
import FlatList from '../lists/FlatList'
import { Grid, Card, CardHeader, IconButton } from '@material-ui/core'
import { formatNotificationRule, sortNotificationRules } from './util'
import { Delete } from '@material-ui/icons'
import UserNotificationRuleDeleteDialog from './UserNotificationRuleDeleteDialog'

const query = gql`
  query nrList($id: ID!) {
    user(id: $id) {
      id
      notificationRules {
        id
        delayMinutes
        contactMethod {
          id
          type
          name
          value
        }
      }
    }
  }
`

export default class UserNotificationRuleList extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,
    readOnly: p.bool,
  }

  state = {
    delete: null,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.userID }}
        render={({ data }) => this.renderList(data.user.notificationRules)}
      />
    )
  }

  renderList(notificationRules) {
    return (
      <Grid item xs={12}>
        <Card>
          <CardHeader title='Notification Rules' />
          <FlatList
            data-cy='notification-rules'
            items={sortNotificationRules(notificationRules).map(nr => ({
              title: formatNotificationRule(nr.delayMinutes, nr.contactMethod),
              action: this.props.readOnly ? null : (
                <IconButton
                  aria-label='Delete notification rule'
                  onClick={() => this.setState({ delete: nr.id })}
                >
                  <Delete />
                </IconButton>
              ),
            }))}
            emptyMessage='No notification rules'
          />
        </Card>
        {this.state.delete && (
          <UserNotificationRuleDeleteDialog
            ruleID={this.state.delete}
            onClose={() => this.setState({ delete: null })}
          />
        )}
      </Grid>
    )
  }
}
