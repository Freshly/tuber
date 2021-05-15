import { gql } from '@apollo/client'

export const GET_APPS = gql`
    query GetApps {
        getApps {
            name
        }
    }
`
