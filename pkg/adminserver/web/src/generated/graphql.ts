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
  githubRepo?: Maybe<Scalars['String']>;
  slackChannel?: Maybe<Scalars['String']>;
  cloudSourceRepo?: Maybe<Scalars['String']>;
};

export type ClusterInfo = {
  __typename?: 'ClusterInfo';
  name: Scalars['String'];
  region: Scalars['String'];
};

export type CreateReviewAppInput = {
  name: Scalars['String'];
  branchName: Scalars['String'];
};

export type ManualApplyInput = {
  name: Scalars['ID'];
  resources: Array<Maybe<Scalars['String']>>;
};

export type Mutation = {
  __typename?: 'Mutation';
  createApp?: Maybe<TuberApp>;
  updateApp?: Maybe<TuberApp>;
  removeApp?: Maybe<TuberApp>;
  deploy?: Maybe<TuberApp>;
  destroyApp?: Maybe<TuberApp>;
  createReviewApp?: Maybe<TuberApp>;
  setAppVar?: Maybe<TuberApp>;
  unsetAppVar?: Maybe<TuberApp>;
  setAppEnv?: Maybe<TuberApp>;
  unsetAppEnv?: Maybe<TuberApp>;
  setExcludedResource?: Maybe<TuberApp>;
  unsetExcludedResource?: Maybe<TuberApp>;
  rollback?: Maybe<TuberApp>;
  setGithubRepo?: Maybe<TuberApp>;
  setCloudSourceRepo?: Maybe<TuberApp>;
  setSlackChannel?: Maybe<TuberApp>;
  manualApply?: Maybe<TuberApp>;
  setRacEnabled?: Maybe<TuberApp>;
  setRacVar?: Maybe<TuberApp>;
  unsetRacVar?: Maybe<TuberApp>;
  setRacExclusion?: Maybe<TuberApp>;
  unsetRacExclusion?: Maybe<TuberApp>;
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


export type MutationDeployArgs = {
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


export type MutationSetExcludedResourceArgs = {
  input: SetResourceInput;
};


export type MutationUnsetExcludedResourceArgs = {
  input: SetResourceInput;
};


export type MutationRollbackArgs = {
  input: AppInput;
};


export type MutationSetGithubRepoArgs = {
  input: AppInput;
};


export type MutationSetCloudSourceRepoArgs = {
  input: AppInput;
};


export type MutationSetSlackChannelArgs = {
  input: AppInput;
};


export type MutationManualApplyArgs = {
  input: ManualApplyInput;
};


export type MutationSetRacEnabledArgs = {
  input: SetRacEnabledInput;
};


export type MutationSetRacVarArgs = {
  input: SetTupleInput;
};


export type MutationUnsetRacVarArgs = {
  input: SetTupleInput;
};


export type MutationSetRacExclusionArgs = {
  input: SetResourceInput;
};


export type MutationUnsetRacExclusionArgs = {
  input: SetResourceInput;
};

export type Query = {
  __typename?: 'Query';
  getApp?: Maybe<TuberApp>;
  getApps: Array<TuberApp>;
  getClusterInfo: ClusterInfo;
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

export type SetRacEnabledInput = {
  name: Scalars['ID'];
  enabled: Scalars['Boolean'];
};

export type SetResourceInput = {
  appName: Scalars['ID'];
  name: Scalars['String'];
  kind: Scalars['String'];
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
  currentTags?: Maybe<Array<Scalars['String']>>;
  githubRepo: Scalars['String'];
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

export type DeployMutationVariables = Exact<{
  input: AppInput;
}>;


export type DeployMutation = (
  { __typename?: 'Mutation' }
  & { deploy?: Maybe<(
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

export type GetClusterInfoQueryVariables = Exact<{ [key: string]: never; }>;


export type GetClusterInfoQuery = (
  { __typename?: 'Query' }
  & { getClusterInfo: (
    { __typename?: 'ClusterInfo' }
    & Pick<ClusterInfo, 'name' | 'region'>
  ) }
);

export type GetFullAppQueryVariables = Exact<{
  name: Scalars['String'];
}>;


export type GetFullAppQuery = (
  { __typename?: 'Query' }
  & { getApp?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name' | 'reviewApp' | 'cloudSourceRepo' | 'githubRepo' | 'slackChannel' | 'paused' | 'imageTag'>
    & { reviewAppsConfig?: Maybe<(
      { __typename?: 'ReviewAppsConfig' }
      & Pick<ReviewAppsConfig, 'enabled'>
      & { excludedResources: Array<(
        { __typename?: 'Resource' }
        & Pick<Resource, 'kind' | 'name'>
      )>, vars: Array<(
        { __typename?: 'Tuple' }
        & Pick<Tuple, 'key' | 'value'>
      )> }
    )>, vars: Array<(
      { __typename?: 'Tuple' }
      & Pick<Tuple, 'key' | 'value'>
    )>, env?: Maybe<Array<(
      { __typename?: 'Tuple' }
      & Pick<Tuple, 'key' | 'value'>
    )>>, reviewApps?: Maybe<Array<(
      { __typename?: 'TuberApp' }
      & Pick<TuberApp, 'name'>
    )>>, excludedResources: Array<(
      { __typename?: 'Resource' }
      & Pick<Resource, 'name' | 'kind'>
    )> }
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

export type SetCloudSourceRepoMutationVariables = Exact<{
  input: AppInput;
}>;


export type SetCloudSourceRepoMutation = (
  { __typename?: 'Mutation' }
  & { setCloudSourceRepo?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name' | 'cloudSourceRepo'>
  )> }
);

export type SetExcludedResourceMutationVariables = Exact<{
  input: SetResourceInput;
}>;


export type SetExcludedResourceMutation = (
  { __typename?: 'Mutation' }
  & { setExcludedResource?: Maybe<(
    { __typename?: 'TuberApp' }
    & { excludedResources: Array<(
      { __typename?: 'Resource' }
      & Pick<Resource, 'name' | 'kind'>
    )> }
  )> }
);

export type SetGithubRepoMutationVariables = Exact<{
  input: AppInput;
}>;


export type SetGithubRepoMutation = (
  { __typename?: 'Mutation' }
  & { setGithubRepo?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name' | 'githubRepo'>
  )> }
);

export type SetRacEnabledMutationVariables = Exact<{
  input: SetRacEnabledInput;
}>;


export type SetRacEnabledMutation = (
  { __typename?: 'Mutation' }
  & { setRacEnabled?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
    & { reviewAppsConfig?: Maybe<(
      { __typename?: 'ReviewAppsConfig' }
      & Pick<ReviewAppsConfig, 'enabled'>
    )> }
  )> }
);

export type SetRacExclusionMutationVariables = Exact<{
  input: SetResourceInput;
}>;


export type SetRacExclusionMutation = (
  { __typename?: 'Mutation' }
  & { setRacExclusion?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
    & { reviewAppsConfig?: Maybe<(
      { __typename?: 'ReviewAppsConfig' }
      & { excludedResources: Array<(
        { __typename?: 'Resource' }
        & Pick<Resource, 'name' | 'kind'>
      )> }
    )> }
  )> }
);

export type SetRacVarMutationVariables = Exact<{
  input: SetTupleInput;
}>;


export type SetRacVarMutation = (
  { __typename?: 'Mutation' }
  & { setRacVar?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
    & { reviewAppsConfig?: Maybe<(
      { __typename?: 'ReviewAppsConfig' }
      & { vars: Array<(
        { __typename?: 'Tuple' }
        & Pick<Tuple, 'key' | 'value'>
      )> }
    )> }
  )> }
);

export type SetSlackChannelMutationVariables = Exact<{
  input: AppInput;
}>;


export type SetSlackChannelMutation = (
  { __typename?: 'Mutation' }
  & { setSlackChannel?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name' | 'slackChannel'>
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

export type UnsetExcludedResourceMutationVariables = Exact<{
  input: SetResourceInput;
}>;


export type UnsetExcludedResourceMutation = (
  { __typename?: 'Mutation' }
  & { unsetExcludedResource?: Maybe<(
    { __typename?: 'TuberApp' }
    & { excludedResources: Array<(
      { __typename?: 'Resource' }
      & Pick<Resource, 'name' | 'kind'>
    )> }
  )> }
);

export type UnsetRacExclusionMutationVariables = Exact<{
  input: SetResourceInput;
}>;


export type UnsetRacExclusionMutation = (
  { __typename?: 'Mutation' }
  & { unsetRacExclusion?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
    & { reviewAppsConfig?: Maybe<(
      { __typename?: 'ReviewAppsConfig' }
      & { excludedResources: Array<(
        { __typename?: 'Resource' }
        & Pick<Resource, 'name' | 'kind'>
      )> }
    )> }
  )> }
);

export type UnsetRacVarMutationVariables = Exact<{
  input: SetTupleInput;
}>;


export type UnsetRacVarMutation = (
  { __typename?: 'Mutation' }
  & { unsetRacVar?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name'>
    & { reviewAppsConfig?: Maybe<(
      { __typename?: 'ReviewAppsConfig' }
      & { vars: Array<(
        { __typename?: 'Tuple' }
        & Pick<Tuple, 'key' | 'value'>
      )> }
    )> }
  )> }
);

export type UpdateAppMutationVariables = Exact<{
  input: AppInput;
}>;


export type UpdateAppMutation = (
  { __typename?: 'Mutation' }
  & { updateApp?: Maybe<(
    { __typename?: 'TuberApp' }
    & Pick<TuberApp, 'name' | 'paused' | 'slackChannel' | 'githubRepo' | 'cloudSourceRepo' | 'imageTag'>
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
        "name": "ClusterInfo",
        "fields": [
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
            "name": "region",
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
            "name": "deploy",
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
            "name": "setExcludedResource",
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
            "name": "unsetExcludedResource",
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
            "name": "rollback",
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
            "name": "setGithubRepo",
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
            "name": "setCloudSourceRepo",
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
            "name": "setSlackChannel",
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
            "name": "manualApply",
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
            "name": "setRacEnabled",
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
            "name": "setRacVar",
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
            "name": "unsetRacVar",
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
            "name": "setRacExclusion",
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
            "name": "unsetRacExclusion",
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
          },
          {
            "name": "getClusterInfo",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "ClusterInfo",
                "ofType": null
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
            "name": "currentTags",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "SCALAR",
                  "name": "Any"
                }
              }
            },
            "args": []
          },
          {
            "name": "githubRepo",
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
export const DeployDocument = gql`
    mutation Deploy($input: AppInput!) {
  deploy(input: $input) {
    name
  }
}
    `;

export function useDeployMutation() {
  return Urql.useMutation<DeployMutation, DeployMutationVariables>(DeployDocument);
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
export const GetClusterInfoDocument = gql`
    query GetClusterInfo {
  getClusterInfo {
    name
    region
  }
}
    `;

export function useGetClusterInfoQuery(options: Omit<Urql.UseQueryArgs<GetClusterInfoQueryVariables>, 'query'> = {}) {
  return Urql.useQuery<GetClusterInfoQuery>({ query: GetClusterInfoDocument, ...options });
};
export const GetFullAppDocument = gql`
    query GetFullApp($name: String!) {
  getApp(name: $name) {
    name
    reviewApp
    cloudSourceRepo
    githubRepo
    slackChannel
    paused
    imageTag
    reviewAppsConfig {
      enabled
      excludedResources {
        kind
        name
      }
      vars {
        key
        value
      }
    }
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
    excludedResources {
      name
      kind
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
export const SetCloudSourceRepoDocument = gql`
    mutation SetCloudSourceRepo($input: AppInput!) {
  setCloudSourceRepo(input: $input) {
    name
    cloudSourceRepo
  }
}
    `;

export function useSetCloudSourceRepoMutation() {
  return Urql.useMutation<SetCloudSourceRepoMutation, SetCloudSourceRepoMutationVariables>(SetCloudSourceRepoDocument);
};
export const SetExcludedResourceDocument = gql`
    mutation SetExcludedResource($input: SetResourceInput!) {
  setExcludedResource(input: $input) {
    excludedResources {
      name
      kind
    }
  }
}
    `;

export function useSetExcludedResourceMutation() {
  return Urql.useMutation<SetExcludedResourceMutation, SetExcludedResourceMutationVariables>(SetExcludedResourceDocument);
};
export const SetGithubRepoDocument = gql`
    mutation SetGithubRepo($input: AppInput!) {
  setGithubRepo(input: $input) {
    name
    githubRepo
  }
}
    `;

export function useSetGithubRepoMutation() {
  return Urql.useMutation<SetGithubRepoMutation, SetGithubRepoMutationVariables>(SetGithubRepoDocument);
};
export const SetRacEnabledDocument = gql`
    mutation SetRacEnabled($input: SetRacEnabledInput!) {
  setRacEnabled(input: $input) {
    name
    reviewAppsConfig {
      enabled
    }
  }
}
    `;

export function useSetRacEnabledMutation() {
  return Urql.useMutation<SetRacEnabledMutation, SetRacEnabledMutationVariables>(SetRacEnabledDocument);
};
export const SetRacExclusionDocument = gql`
    mutation SetRacExclusion($input: SetResourceInput!) {
  setRacExclusion(input: $input) {
    name
    reviewAppsConfig {
      excludedResources {
        name
        kind
      }
    }
  }
}
    `;

export function useSetRacExclusionMutation() {
  return Urql.useMutation<SetRacExclusionMutation, SetRacExclusionMutationVariables>(SetRacExclusionDocument);
};
export const SetRacVarDocument = gql`
    mutation SetRacVar($input: SetTupleInput!) {
  setRacVar(input: $input) {
    name
    reviewAppsConfig {
      vars {
        key
        value
      }
    }
  }
}
    `;

export function useSetRacVarMutation() {
  return Urql.useMutation<SetRacVarMutation, SetRacVarMutationVariables>(SetRacVarDocument);
};
export const SetSlackChannelDocument = gql`
    mutation SetSlackChannel($input: AppInput!) {
  setSlackChannel(input: $input) {
    name
    slackChannel
  }
}
    `;

export function useSetSlackChannelMutation() {
  return Urql.useMutation<SetSlackChannelMutation, SetSlackChannelMutationVariables>(SetSlackChannelDocument);
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
export const UnsetExcludedResourceDocument = gql`
    mutation UnsetExcludedResource($input: SetResourceInput!) {
  unsetExcludedResource(input: $input) {
    excludedResources {
      name
      kind
    }
  }
}
    `;

export function useUnsetExcludedResourceMutation() {
  return Urql.useMutation<UnsetExcludedResourceMutation, UnsetExcludedResourceMutationVariables>(UnsetExcludedResourceDocument);
};
export const UnsetRacExclusionDocument = gql`
    mutation UnsetRacExclusion($input: SetResourceInput!) {
  unsetRacExclusion(input: $input) {
    name
    reviewAppsConfig {
      excludedResources {
        name
        kind
      }
    }
  }
}
    `;

export function useUnsetRacExclusionMutation() {
  return Urql.useMutation<UnsetRacExclusionMutation, UnsetRacExclusionMutationVariables>(UnsetRacExclusionDocument);
};
export const UnsetRacVarDocument = gql`
    mutation UnsetRacVar($input: SetTupleInput!) {
  unsetRacVar(input: $input) {
    name
    reviewAppsConfig {
      vars {
        key
        value
      }
    }
  }
}
    `;

export function useUnsetRacVarMutation() {
  return Urql.useMutation<UnsetRacVarMutation, UnsetRacVarMutationVariables>(UnsetRacVarDocument);
};
export const UpdateAppDocument = gql`
    mutation UpdateApp($input: AppInput!) {
  updateApp(input: $input) {
    name
    paused
    slackChannel
    githubRepo
    cloudSourceRepo
    imageTag
  }
}
    `;

export function useUpdateAppMutation() {
  return Urql.useMutation<UpdateAppMutation, UpdateAppMutationVariables>(UpdateAppDocument);
};