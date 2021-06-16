/* eslint-disable react/prop-types */
import { useRouter } from 'next/dist/client/router'
import React, { useRef } from 'react'
import { Card, Heading, TextInput, TextInputGroup, ExcludedResources, Collapsible, TextInputForm } from '../../src/components'
import { throwError } from '../../src/throwError'
import { TrashIcon } from '@heroicons/react/outline'
import {
	useDeployMutation,
	useGetFullAppQuery,
	useDestroyAppMutation,
	useCreateReviewAppMutation,
	useSetExcludedResourceMutation, useUnsetExcludedResourceMutation,
	useSetAppVarMutation, useUnsetAppVarMutation,
	useSetAppEnvMutation, useUnsetAppEnvMutation, useSetCloudSourceRepoMutation, useSetSlackChannelMutation, useSetGithubUrlMutation,
} from '../../src/generated/graphql'


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
	const [{ error: destroyAppError }, destroyApp] = useDestroyAppMutation()
	const hostname = `https://${app.name}.staging.freshlyservices.net/`
	const [{ error: deployErr }, deploy] = useDeployMutation()

	return <div>
		<section className="flex justify-between p-3 mb-2">
			<div className="flex justify-between">
				<div className="mr-3">
					<h1 className="text-3xl">{app.name}</h1>
					<div>
						<small>
							<a href={hostname} target="_blank" rel="noreferrer">{hostname}</a>
						</small>
					</div>
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

			<div>
				{deployErr && <div className="bg-red-700 text-white border-red-700 p-2">
					{deployErr.message}
				</div>}

				<div className="text-white bg-green-700 p-2 rounded-md cursor-pointer" onClick={() => deploy({ input: { name: app.name } }) }>
					<span>Deploy</span>
				</div>
			</div>
		</section>

		<section>
			<Card className="mb-2">
				<div className="inline-grid grid-cols-2 leading-7">
					<div>Slack Channel</div>
					<TextInputForm
						value={app.slackChannel}
						keyName="slackChannel"
						appName={app.name}
						useSet={useSetSlackChannelMutation}
					/>

					<div>Github URL</div>
					<TextInputForm
						value={app.githubURL}
						keyName="githubURL"
						appName={app.name}
						useSet={useSetGithubUrlMutation}
					/>

					<div>Cloud Source Repo</div>
					<TextInputForm
						value={app.cloudSourceRepo}
						keyName="cloudSourceRepo"
						appName={app.name}
						useSet={useSetCloudSourceRepoMutation}
					/>
				</div>
			</Card>
		</section>

		<section>
			<Card className="mb-2">
				<h2 className="text-xl mb-2">YAML Interpolation Vars</h2>
				<TextInputGroup
					vars={app.vars} appName={app.name}
					useSet={useSetAppVarMutation}
					useUnset={useUnsetAppVarMutation}
				/>
			</Card>
		</section>

		{app.reviewApp || <>
			<Card className="mb-2">
				<h2 className="text-xl mb-2">Create a review app</h2>
				<CreateForm app={app} />
				<Heading>Review apps</Heading>
				{destroyAppError && <div className="bg-red-700 text-white border-red-700 p-2">
					{destroyAppError.message}
				</div>}

				{app.reviewApps && app.reviewApps.map(reviewApp =>
					<div key={reviewApp.name}>
						<a href={`/tuber/apps/${reviewApp.name}`}>{reviewApp.name}</a>
						<TrashIcon className="w-5" onClick={() => destroyApp({ input: { name: reviewApp.name } })}/>
					</div>,
				)}
			</Card>
		</>}

		<Card className="mb-2">
			<ExcludedResources
				appName={app.name}
				resources={app.excludedResources}
				useSet={useSetExcludedResourceMutation}
				useUnset={useUnsetExcludedResourceMutation}
			/>
		</Card>

		<Card>
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
