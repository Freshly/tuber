/* eslint-disable react/prop-types */
import { useRouter } from 'next/dist/client/router'
import React, { useRef } from 'react'
import Switch from 'react-switch'
import dateformat from 'dateformat'
import { Card, Heading, TextInput, TextInputGroup, ExcludedResources, Collapsible, TextInputForm, ConfirmButton, Button } from '../../src/components'
import { throwError } from '../../src/throwError'
import { TrashIcon } from '@heroicons/react/outline'
import {
	useDeployMutation,
	useUpdateAppMutation,
	useGetFullAppQuery,
	useDestroyAppMutation,
	useCreateReviewAppMutation,
	useSetRacEnabledMutation,
	useSetRacExclusionMutation, useUnsetRacExclusionMutation,
	useSetRacVarMutation, useUnsetRacVarMutation,
	useSetExcludedResourceMutation, useUnsetExcludedResourceMutation,
	useSetAppVarMutation, useUnsetAppVarMutation,
	useSetAppEnvMutation, useUnsetAppEnvMutation, useSetCloudSourceRepoMutation, useSetSlackChannelMutation, useSetGithubRepoMutation, useGetClusterInfoQuery,
} from '../../src/generated/graphql'
import Head from 'next/head'
import { useClusterInfo } from '../../src/useClusterInfo'

const CreateForm = ({ app }) => {
	const [{ error, fetching }, create] = useCreateReviewAppMutation()
	const branchNameRef = useRef(null)

	const submit = (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()

		create({
			input: {
				name:       app.name,
				branchName: branchNameRef.current.value,
			},
		})
	}

	return <form onSubmit={submit}>
		{error && <div className="bg-red-700 text-white border-red-700 p-2">
			{error.message}
		</div>}
		<TextInput name="branchName" ref={branchNameRef} placeholder="branch name" required disabled={fetching} />
		<button type="submit" className="rounded-sm p-1 underline" disabled={fetching}>Create</button>
	</form>
}

const ShowApp = () => {
	const router = useRouter()
	const id = router.query.id as string
	const [{ data: { getApp: app } }] = throwError(useGetFullAppQuery({ variables: { name: id } }))
	const clusterInfo = useClusterInfo()
	const [{ error: destroyAppError }, destroyApp] = useDestroyAppMutation()
	const [, setEnabled] = useSetRacEnabledMutation()
	const hostname = `https://${app.name}.${clusterInfo.name}.freshlyservices.net/`

	return <div>
		<Head>
			<title>{`${app.name} - ${clusterInfo.name}`}</title>
		</Head>

		<section className="flex justify-between p-3 mb-2">
			<div className="flex justify-between">
				<div className="mr-3">
					<h1 className="text-3xl">{app.name}</h1>
					<div>
						<small>
							<a href="https://app.datadoghq.com/apm/home?env=production" target="_blank" rel="noreferrer">DataDog Logs</a>
						</small>
					</div>
					<div>
						<small>
							<a href="https://console.cloud.google.com/" target="_blank" rel="noreferrer">GKE Dashboard</a>
						</small>
					</div>
				</div>
			</div>

			<div className="flex">
				<div className="mr-1">
					<Button
						input={{ name: app.name, paused: !app.paused }}
						title={app.paused ? 'Resume' : 'Pause'}
						useMutation={useUpdateAppMutation}
						className="bg-yellow-700 border-yellow-700"
					/>
				</div>
				<ConfirmButton input={{ name: app.name }} title={'Deploy'} useMutation={useDeployMutation} className="bg-green-700 border-green-700" />
			</div>
		</section>

		<section>
			<Card>
				<div
					className="inline-grid leading-8"
					style={{ 'gridTemplateColumns': 'repeat(2, minmax(300px, 352px))' }}
				>
					<div>Cluster Link</div>
					<div><a href={hostname} target="_blank" rel="noreferrer" className="underline whitespace-nowrap">{hostname}</a></div>

					<div>Slack Channel</div>
					<TextInputForm
						value={app.slackChannel}
						useSet={useSetSlackChannelMutation}
						appName={app.name}
						keyName="slackChannel"
						className="min-w-300px"
						required={false}
					/>

					<div>Github Repo</div>
					<TextInputForm
						value={app.githubRepo}
						useSet={useSetGithubRepoMutation}
						appName={app.name}
						keyName="githubRepo"
						className="min-w-300px"
					/>

					<div>Cloud Source Repo</div>
					<TextInputForm
						value={app.cloudSourceRepo}
						useSet={useSetCloudSourceRepoMutation}
						appName={app.name}
						keyName="cloudSourceRepo"
						className="min-w-300px"
					/>

					<div>Image Tag</div>
					<TextInputForm
						value={app.imageTag}
						useSet={useUpdateAppMutation}
						appName={app.name}
						keyName="imageTag"
						className="min-w-300px"
					/>
				</div>
			</Card>
		</section>
		<section>
			<Card>
				<h2 className="text-xl mb-2">Recent Builds</h2>
				{app.cloudBuildStatuses.map(bs =>
					<div key={bs.startTime} className="grid grid-cols-3">
						<span>{dateformat(bs.startTime, 'ddd mmm dS, h:MM TT')}</span>
						<span>{bs.status}</span>
						<a href={bs.Link} className="underline" target="_blank" rel="noreferrer">Build Logs</a>
					</div>,
				)}
			</Card>
		</section>

		<section>
			<Card>
				<h2 className="text-xl mb-2">YAML Interpolation Vars</h2>
				<TextInputGroup
					vars={app.vars} appName={app.name}
					useSet={useSetAppVarMutation}
					useUnset={useUnsetAppVarMutation}
				/>
			</Card>
		</section>

		{(clusterInfo.reviewAppsEnabled && !app.reviewApp) && <>
			<Card>
				<h2 className="text-xl mb-2">Create a review app</h2>
				<CreateForm app={app} />
				{destroyAppError && <div className="bg-red-700 text-white border-red-700 p-2">
					{destroyAppError.message}
				</div>}

				{app.reviewApps && <Heading>Review apps</Heading>}
				{app.reviewApps && app.reviewApps.map(reviewApp =>
					<div key={reviewApp.name}>
						<a href={`/tuber/apps/${reviewApp.name}`}>{reviewApp.name}</a>
						<TrashIcon className="w-5" onClick={() => destroyApp({ input: { name: reviewApp.name } })}/>
					</div>,
				)}
			</Card>

			<Card>
				<div className="mb-4">
					<h2 className="text-xl">Configure Review Apps</h2>
					<p className=""><small>Configure how review apps created based off this app behave</small></p>
				</div>

				<div className="mb-4">
					<label>
						<div className="mb-2">Enable/Disable Review Apps</div>
						<Switch
							onChange={() => { setEnabled({ input: { name: app.name, enabled: !app.reviewAppsConfig.enabled } }) }}
							checked={app.reviewAppsConfig.enabled}
						/>
					</label>
				</div>

				<div className="mb-4">
					<h3>Review App Vars</h3>
					<TextInputGroup
						vars={app.reviewAppsConfig.vars} appName={app.name}
						useSet={useSetRacVarMutation}
						useUnset={useUnsetRacVarMutation}
					/>
				</div>

				<div className="mb-4">
					<h3 className="text-l mb-2">Excluded Resources</h3>
					<ExcludedResources
						appName={app.name}
						resources={app.reviewAppsConfig.excludedResources}
						useSet={useSetRacExclusionMutation}
						useUnset={useUnsetRacExclusionMutation}
					/>
				</div>
			</Card>
		</>}

		<Card>
			<h2 className="text-xl mb-2">Excluded Resources</h2>
			<ExcludedResources
				appName={app.name}
				resources={app.excludedResources}
				useSet={useSetExcludedResourceMutation}
				useUnset={useUnsetExcludedResourceMutation}
			/>
		</Card>

		<Card className="shadow-dark-50 shadow">
			<Collapsible heading={'Environment Variables'} collapsed={true}>
				<TextInputGroup
					vars={app.env} appName={app.name}
					useSet={useSetAppEnvMutation}
					useUnset={useUnsetAppEnvMutation}
				/>
			</Collapsible>
		</Card>
	</div>
}

export default ShowApp
