import gql from 'graphql-tag';
import * as Urql from 'urql';
export type Maybe<T> = T | null;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
};


export type AppInput = {
  name: Scalars['ID'];
  isIstio?: Maybe<Scalars['Boolean']>;
  imageTag?: Maybe<Scalars['String']>;
  paused?: Maybe<Scalars['Boolean']>;
};

export type CreateReviewAppInput = {
  name: Scalars['String'];
  branchName: Scalars['String'];
};

export type Mutation = {
  __typename?: 'Mutation';
  createApp?: Maybe<TuberApp>;
  updateApp?: Maybe<TuberApp>;
  removeApp?: Maybe<TuberApp>;
  destroyApp?: Maybe<TuberApp>;
  createReviewApp?: Maybe<TuberApp>;
  setAppVar?: Maybe<TuberApp>;
  unsetAppVar?: Maybe<TuberApp>;
  setAppEnv?: Maybe<TuberApp>;
  unsetAppEnv?: Maybe<TuberApp>;
  excludedResources: Array<Resource>;
};


export type MutationCreateAppArgs = {
  input: AppInput;
};


export type MutationUpdateAppArgs = {
  input: AppInput;
};


export type MutationRemoveAppArgs = {
  input: AppInput;
};


export type MutationDestroyAppArgs = {
  input: AppInput;
};


export type MutationCreateReviewAppArgs = {
  input: CreateReviewAppInput;
};


export type MutationSetAppVarArgs = {
  input: SetTupleInput;
};


export type MutationUnsetAppVarArgs = {
  input: SetTupleInput;
};


export type MutationSetAppEnvArgs = {
  input: SetTupleInput;
};


export type MutationUnsetAppEnvArgs = {
  input: SetTupleInput;
};

export type Query = {
  __typename?: 'Query';
  getApp?: Maybe<TuberApp>;
  getApps: Array<TuberApp>;
};


export type QueryGetAppArgs = {
  name: Scalars['String'];
};

export type Resource = {
  __typename?: 'Resource';
  encoded: Scalars['String'];
  kind: Scalars['String'];
  name: Scalars['String'];
};

export type ReviewAppsConfig = {
  __typename?: 'ReviewAppsConfig';
  enabled: Scalars['Boolean'];
  vars: Array<Tuple>;
  excludedResources: Array<Resource>;
};

export type SetTupleInput = {
  name: Scalars['ID'];
  key: Scalars['String'];
  value: Scalars['String'];
};

export type State = {
  __typename?: 'State';
  Current: Array<Resource>;
  Previous: Array<Resource>;
};

export type TuberApp = {
  __typename?: 'TuberApp';
  cloudSourceRepo: Scalars['String'];
  imageTag: Scalars['String'];
  name: Scalars['ID'];
  paused: Scalars['Boolean'];
  reviewApp: Scalars['Boolean'];
  reviewAppsConfig?: Maybe<ReviewAppsConfig>;
  slackChannel: Scalars['String'];
  sourceAppName: Scalars['String'];
  state: State;
  triggerID: Scalars['String'];
  vars: Array<Tuple>;
  reviewApps?: Maybe<Array<TuberApp>>;
  env?: Maybe<Array<Tuple>>;
  excludedResources: Array<Resource>;
};

export type Tuple = {
  __typename?: 'Tuple';
  key: Scalars['String'];
  value: Scalars['String'];
};

export type CreateReviewAppMutationVariables = Exact<{
  input: CreateReviewAppInput;
}>;


export type CreateReviewAppMutation = (
  { __typename?: 'Mutation' }
  & { createReviewApp?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
  )> }
);

export type DestroyAppMutationVariables = Exact<{
  input: AppInput;
}>;


export type DestroyAppMutation = (
  { __typename?: 'Mutation' }
  & { destroyApp?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
  )> }
);

export type GetAppQueryVariables = Exact<{
  name: Scalars['String'];
}>;


export type GetAppQuery = (
  { __typename?: 'Query' }
  & { getApp?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
  )> }
);

export type GetAppsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetAppsQuery = (
  { __typename?: 'Query' }
  & { getApps: Array<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name' | 'paused' | 'imageTag'>
  )> }
);

export type GetFullAppQueryVariables = Exact<{
  name: Scalars['String'];
}>;


export type GetFullAppQuery = (
  { __typename?: 'Query' }
  & { getApp?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name' | 'reviewApp'>
    & { vars: Array<(
      { __typename?: 'Tuple' }
      & Pick<Tuple, 'key' | 'value'>
    )>, env?: Maybe<Array<(
      { __typename?: 'Tuple' }
      & Pick<Tuple, 'key' | 'value'>
    )>>, reviewApps?: Maybe<Array<(
      { __typename?: 'TuberApp' }
      & Pick<TuberApp, 'name'>
    )>> }
  )> }
);

export type SetAppEnvMutationVariables = Exact<{
  input: SetTupleInput;
}>;


export type SetAppEnvMutation = (
  { __typename?: 'Mutation' }
  & { setAppEnv?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
    & { env?: Maybe<Array<(
      { __typename?: 'Tuple' }
      & Pick<Tuple, 'key' | 'value'>
    )>> }
  )> }
);

export type SetAppVarMutationVariables = Exact<{
  input: SetTupleInput;
}>;


export type SetAppVarMutation = (
  { __typename?: 'Mutation' }
  & { setAppVar?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
    & { vars: Array<(
      { __typename?: 'Tuple' }
      & Pick<Tuple, 'key' | 'value'>
    )> }
  )> }
);

export type UnsetAppEnvMutationVariables = Exact<{
  input: SetTupleInput;
}>;


export type UnsetAppEnvMutation = (
  { __typename?: 'Mutation' }
  & { unsetAppEnv?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
    & { env?: Maybe<Array<(
      { __typename?: 'Tuple' }
      & Pick<Tuple, 'key' | 'value'>
    )>> }
  )> }
);

export type UnsetAppVarMutationVariables = Exact<{
  input: SetTupleInput;
}>;


export type UnsetAppVarMutation = (
  { __typename?: 'Mutation' }
  & { unsetAppVar?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
    & { env?: Maybe<Array<(
      { __typename?: 'Tuple' }
      & Pick<Tuple, 'key' | 'value'>
    )>> }
  )> }
);

import { IntrospectionQuery } from 'graphql';
export default {
  "__schema": {
    "queryType": {
      "name": "Query"
    },
    "mutationType": {
      "name": "Mutation"
    },
    "subscriptionType": null,
    "types": [
      {
        "kind": "OBJECT",
        "name": "Mutation",
        "fields": [
          {
            "name": "createApp",
            "type": {
              "kind": "OBJECT",
              "name": "TuberApp",
              "ofType": null
            },
            "args": [
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "updateApp",
            "type": {
              "kind": "OBJECT",
              "name": "TuberApp",
              "ofType": null
            },
            "args": [
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "removeApp",
            "type": {
              "kind": "OBJECT",
              "name": "TuberApp",
              "ofType": null
            },
            "args": [
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "destroyApp",
            "type": {
              "kind": "OBJECT",
              "name": "TuberApp",
              "ofType": null
            },
            "args": [
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "createReviewApp",
            "type": {
              "kind": "OBJECT",
              "name": "TuberApp",
              "ofType": null
            },
            "args": [
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "setAppVar",
            "type": {
              "kind": "OBJECT",
              "name": "TuberApp",
              "ofType": null
            },
            "args": [
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "unsetAppVar",
            "type": {
              "kind": "OBJECT",
              "name": "TuberApp",
              "ofType": null
            },
            "args": [
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "setAppEnv",
            "type": {
              "kind": "OBJECT",
              "name": "TuberApp",
              "ofType": null
            },
            "args": [
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "unsetAppEnv",
            "type": {
              "kind": "OBJECT",
              "name": "TuberApp",
              "ofType": null
            },
            "args": [
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "excludedResources",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "LIST",
                "ofType": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "OBJECT",
                    "name": "Resource",
                    "ofType": null
                  }
                }
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Query",
        "fields": [
          {
            "name": "getApp",
            "type": {
              "kind": "OBJECT",
              "name": "TuberApp",
              "ofType": null
            },
            "args": [
              {
                "name": "name",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "getApps",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "LIST",
                "ofType": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "OBJECT",
                    "name": "TuberApp",
                    "ofType": null
                  }
                }
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Resource",
        "fields": [
          {
            "name": "encoded",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "kind",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "name",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "ReviewAppsConfig",
        "fields": [
          {
            "name": "enabled",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "vars",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "LIST",
                "ofType": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "OBJECT",
                    "name": "Tuple",
                    "ofType": null
                  }
                }
              }
            },
            "args": []
          },
          {
            "name": "excludedResources",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "LIST",
                "ofType": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "OBJECT",
                    "name": "Resource",
                    "ofType": null
                  }
                }
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "State",
        "fields": [
          {
            "name": "Current",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "LIST",
                "ofType": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "OBJECT",
                    "name": "Resource",
                    "ofType": null
                  }
                }
              }
            },
            "args": []
          },
          {
            "name": "Previous",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "LIST",
                "ofType": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "OBJECT",
                    "name": "Resource",
                    "ofType": null
                  }
                }
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "TuberApp",
        "fields": [
          {
            "name": "cloudSourceRepo",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "imageTag",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "name",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "paused",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "reviewApp",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "reviewAppsConfig",
            "type": {
              "kind": "OBJECT",
              "name": "ReviewAppsConfig",
              "ofType": null
            },
            "args": []
          },
          {
            "name": "slackChannel",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "sourceAppName",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "state",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "State",
                "ofType": null
              }
            },
            "args": []
          },
          {
            "name": "triggerID",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "vars",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "LIST",
                "ofType": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "OBJECT",
                    "name": "Tuple",
                    "ofType": null
                  }
                }
              }
            },
            "args": []
          },
          {
            "name": "reviewApps",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "TuberApp",
                  "ofType": null
                }
              }
            },
            "args": []
          },
          {
            "name": "env",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "Tuple",
                  "ofType": null
                }
              }
            },
            "args": []
          },
          {
            "name": "excludedResources",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "LIST",
                "ofType": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "OBJECT",
                    "name": "Resource",
                    "ofType": null
                  }
                }
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Tuple",
        "fields": [
          {
            "name": "key",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "value",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "SCALAR",
        "name": "Any"
      }
    ],
    "directives": []
  }
} as unknown as IntrospectionQuery;

export const CreateReviewAppDocument = gql`
    mutation CreateReviewApp($input: CreateReviewAppInput!) {
  createReviewApp(input: $input) {
    name
  }
}
    `;

export function useCreateReviewAppMutation() {
  return Urql.useMutation<CreateReviewAppMutation, CreateReviewAppMutationVariables>(CreateReviewAppDocument);
};
export const DestroyAppDocument = gql`
    mutation DestroyApp($input: AppInput!) {
  destroyApp(input: $input) {
    name
  }
}
    `;

export function useDestroyAppMutation() {
  return Urql.useMutation<DestroyAppMutation, DestroyAppMutationVariables>(DestroyAppDocument);
};
export const GetAppDocument = gql`
    query GetApp($name: String!) {
  getApp(name: $name) {
    name
  }
}
    `;

export function useGetAppQuery(options: Omit<Urql.UseQueryArgs<GetAppQueryVariables>, 'query'> = {}) {
  return Urql.useQuery<GetAppQuery>({ query: GetAppDocument, ...options });
};
export const GetAppsDocument = gql`
    query GetApps {
  getApps {
    name
    paused
    imageTag
  }
}
    `;

export function useGetAppsQuery(options: Omit<Urql.UseQueryArgs<GetAppsQueryVariables>, 'query'> = {}) {
  return Urql.useQuery<GetAppsQuery>({ query: GetAppsDocument, ...options });
};
export const GetFullAppDocument = gql`
    query GetFullApp($name: String!) {
  getApp(name: $name) {
    name
    reviewApp
    vars {
      key
      value
    }
    env {
      key
      value
    }
    reviewApps {
      name
    }
  }
}
    `;

export function useGetFullAppQuery(options: Omit<Urql.UseQueryArgs<GetFullAppQueryVariables>, 'query'> = {}) {
  return Urql.useQuery<GetFullAppQuery>({ query: GetFullAppDocument, ...options });
};
export const SetAppEnvDocument = gql`
    mutation SetAppEnv($input: SetTupleInput!) {
  setAppEnv(input: $input) {
    name
    env {
      key
      value
    }
  }
}
    `;

export function useSetAppEnvMutation() {
  return Urql.useMutation<SetAppEnvMutation, SetAppEnvMutationVariables>(SetAppEnvDocument);
};
export const SetAppVarDocument = gql`
    mutation SetAppVar($input: SetTupleInput!) {
  setAppVar(input: $input) {
    name
    vars {
      key
      value
    }
  }
}
    `;

export function useSetAppVarMutation() {
  return Urql.useMutation<SetAppVarMutation, SetAppVarMutationVariables>(SetAppVarDocument);
};
export const UnsetAppEnvDocument = gql`
    mutation UnsetAppEnv($input: SetTupleInput!) {
  unsetAppEnv(input: $input) {
    name
    env {
      key
      value
    }
  }
}
    `;

export function useUnsetAppEnvMutation() {
  return Urql.useMutation<UnsetAppEnvMutation, UnsetAppEnvMutationVariables>(UnsetAppEnvDocument);
};
export const UnsetAppVarDocument = gql`
    mutation UnsetAppVar($input: SetTupleInput!) {
  unsetAppVar(input: $input) {
    name
    env {
      key
      value
    }
  }
}
    `;

export function useUnsetAppVarMutation() {
  return Urql.useMutation<UnsetAppVarMutation, UnsetAppVarMutationVariables>(UnsetAppVarDocument);
};