package wechatarticle

import (
	"strings"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

type Evaluation struct {
	Passed           bool     `json:"passed"`
	LoginValid       bool     `json:"login_valid"`
	AccountObserved  bool     `json:"account_observed"`
	ArticlesObserved bool     `json:"articles_observed"`
	DownloadObserved bool     `json:"download_observed"`
	EvidenceObserved bool     `json:"evidence_observed"`
	FailureReasons   []string `json:"failure_reasons,omitempty"`
	Summary          string   `json:"summary,omitempty"`
}

func Evaluate(execution Execution, requireDownload bool) Evaluation {
	result := Evaluation{LoginValid: execution.Login.Valid, AccountObserved: len(execution.Accounts) > 0, ArticlesObserved: len(execution.Articles) > 0, DownloadObserved: execution.Downloaded != nil, EvidenceObserved: len(execution.Result.EvidenceRefs) > 0}
	if !result.LoginValid {
		result.FailureReasons = append(result.FailureReasons, "wechat_article_login_invalid_or_expired")
	}
	if !result.ArticlesObserved {
		result.FailureReasons = append(result.FailureReasons, "wechat_article_list_empty")
	}
	if requireDownload && !result.DownloadObserved {
		result.FailureReasons = append(result.FailureReasons, "wechat_article_download_missing")
	}
	if !result.EvidenceObserved {
		result.FailureReasons = append(result.FailureReasons, "wechat_article_evidence_missing")
	}
	result.Passed = execution.Result.Status == control.VerificationSatisfied && len(result.FailureReasons) == 0
	if result.Passed {
		result.Summary = "wechat article evidence verified"
	} else {
		result.Summary = strings.Join(result.FailureReasons, ",")
	}
	return result
}
