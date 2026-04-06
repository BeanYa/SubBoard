package service

import (
	"submanager/model"

	"gorm.io/gorm"
)

// CollectNodesForUser collects all proxy nodes a user has access to
// through their plan's service groups (subscription sources + agents).
func CollectNodesForUser(db *gorm.DB, userID uint) ([]ParsedNode, error) {
	// Find user and check plan
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.PlanID == nil {
		return nil, nil
	}

	// Get service group IDs from plan
	var psgs []model.PlanServiceGroup
	if err := db.Where("plan_id = ?", *user.PlanID).Find(&psgs).Error; err != nil {
		return nil, err
	}
	if len(psgs) == 0 {
		return nil, nil
	}

	sgIDs := make([]uint, len(psgs))
	for i, psg := range psgs {
		sgIDs[i] = psg.ServiceGroupID
	}

	// Collect subscription source IDs from groups
	var gss []model.GroupSubscriptionSource
	db.Where("service_group_id IN ?", sgIDs).Find(&gss)
	subIDs := make([]uint, len(gss))
	for i, gs := range gss {
		subIDs[i] = gs.SubscriptionSourceID
	}

	// Collect agent IDs from groups
	var gas []model.GroupAgent
	db.Where("service_group_id IN ?", sgIDs).Find(&gas)
	agentIDs := make([]uint, len(gas))
	for i, ga := range gas {
		agentIDs[i] = ga.AgentID
	}

	// If no sources at all, return empty
	if len(subIDs) == 0 && len(agentIDs) == 0 {
		return nil, nil
	}

	// Query NodeCache with OR conditions
	query := db.Model(&model.NodeCache{})
	if len(subIDs) > 0 && len(agentIDs) > 0 {
		query = query.Where(
			"(source_type = ? AND source_id IN ?) OR (source_type = ? AND source_id IN ?)",
			"subscription", subIDs, "agent", agentIDs,
		)
	} else if len(subIDs) > 0 {
		query = query.Where("source_type = ? AND source_id IN ?", "subscription", subIDs)
	} else {
		query = query.Where("source_type = ? AND source_id IN ?", "agent", agentIDs)
	}

	var nodes []model.NodeCache
	if err := query.Find(&nodes).Error; err != nil {
		return nil, err
	}

	// Convert to ParsedNode slice
	result := make([]ParsedNode, 0, len(nodes))
	for _, nc := range nodes {
		result = append(result, nodeCacheToParsedNode(nc))
	}

	return result, nil
}

// nodeCacheToParsedNode converts a NodeCache model back to a ParsedNode.
func nodeCacheToParsedNode(nc model.NodeCache) ParsedNode {
	return ParsedNode{
		Name:     nc.Name,
		Protocol: nc.Protocol,
		Address:  nc.Address,
		Port:     nc.Port,
		Extra:    map[string]any(nc.Extra),
		RawLink:  nc.RawLink,
	}
}
