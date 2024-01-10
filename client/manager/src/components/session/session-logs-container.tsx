import { SessionLogs } from '@/components/session/session-logs';
import { useSessionRetrieval } from '@/components/session/session-retrieval-hook';
import { IApp, ISession } from '@polo/common/state/models';
import { observer } from 'mobx-react-lite';
import React from 'react';

type TProps = {
	app: IApp;
	session: ISession;
	setOverlayBottom: (val: number) => void;
	onSessionFail: () => void;
}

export const SessionLogsContainer = observer((props: TProps) => {
	useSessionRetrieval(props.app.failures.retrieveFailedSession, props.onSessionFail, props.session);

	const setOverlayProportions = (proportions: number) => {
		const percentage = parseInt(`${proportions * 100}`);
		const inversePercentage = 100 - percentage;
		props.setOverlayBottom(inversePercentage);
	};

	return <SessionLogs
		logs={Array.from(props.session.logs.values())}
		onLogsProportionChanged={setOverlayProportions} />;
});